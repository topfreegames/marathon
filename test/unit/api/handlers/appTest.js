import { expect } from '../common'
import uuid from 'uuid'

describe('Handlers', () => {
  describe('Apps Handler', () => {
    describe('GET', () => {
      it('should return 200 and an empty list of apps if there are no apps', async function () {
        await this.app.db.App.destroy({ truncate: true })
        const res = await this.request.get('/apps')
        expect(res.status).to.equal(200)

        const body = res.body
        expect(body).to.be.an('object')

        expect(body.apps).to.exist()
        expect(body.apps).to.have.length(0)
      })

      it('should return 200 and a list of apps', async function () {
        const app = {
          key: uuid.v4(),
          bundleId: `com.app.${uuid.v4().split('-')[0]}`,
          createdBy: 'someone@somewhere.com',
        }
        await this.app.db.App.create(app)
        const res = await this.request.get('/apps')
        expect(res.status).to.equal(200)

        const body = res.body
        expect(body).to.be.an('object')

        expect(body.apps).to.exist()
        expect(body.apps).to.have.length.at.least(1)

        const myApp = body.apps.filter(a => a.key === app.key)[0]
        expect(myApp).to.exist()
        expect(myApp.id).to.exist()
        expect(myApp.key).to.equal(app.key)
        expect(myApp.bundleId).to.equal(app.bundleId)
        expect(myApp.createdBy).to.equal(app.createdBy)
      })
    })

    describe('POST', () => {
      let app
      let userEmail

      beforeEach(() => {
        app = {
          key: uuid.v4(),
          bundleId: `com.app.${uuid.v4().split('-')[0]}`,
        }
        userEmail = 'someone@somewhere.com'
      })

      it('should return 201 and the created app', async function () {
        const res = await this.request.post('/apps').send(app).set('user-email', userEmail)
        expect(res.status).to.equal(201)

        const body = res.body
        expect(body).to.be.an('object')

        expect(body.app).to.exist()
        expect(body.app).to.be.an('object')

        expect(body.app.id).to.exist()
        expect(body.app.key).to.equal(app.key)
        expect(body.app.bundleId).to.equal(app.bundleId)
        expect(body.app.createdBy).to.equal(userEmail)

        const dbApp = await this.app.db.App.findById(body.app.id)
        expect(dbApp.key).to.equal(app.key)
        expect(dbApp.bundleId).to.equal(app.bundleId)
        expect(dbApp.createdBy).to.equal(userEmail)
      })

      describe('Should fail if', () => {
        it('app with same key already exists', async function () {
          await this.request.post('/apps').send(app).set('user-email', userEmail)

          const res = await this.request.post('/apps').send(app).set('user-email', userEmail)
          expect(res.status).to.equal(409)
        })

        it('missing user-email header', async function () {
          const res = await this.request.post('/apps').send(app)
          expect(res.status).to.equal(422)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.data).to.exist()
          expect(body.data).to.have.length(1)
          expect(body.data[0]).to.have.property('user-email')
          expect(body.data[0]['user-email']).to.contain('empty')
        })

        const tests = [
          { args: 'key' },
          { args: 'bundleId' },
        ]

        tests.forEach((test) => {
          it(`missing ${test.args}`, async function () {
            delete app[test.args]
            const res = await this.request.post('/apps').send(app).set('user-email', userEmail)
            expect(res.status).to.equal(422)

            const body = res.body
            expect(body).to.be.an('object')

            expect(body.data).to.exist()
            expect(body.data).to.have.length(1)
            expect(body.data[0]).to.have.property(test.args)
            expect(body.data[0][test.args]).to.contain('empty')
          })
        })
      })

      it('invalid user-email header', async function () {
        const res = await this.request.post('/apps').send(app).set('user-email', 'not an email')
        expect(res.status).to.equal(422)

        const body = res.body
        expect(body).to.be.an('object')

        expect(body.data).to.exist()
        expect(body.data).to.have.length(1)
        expect(body.data[0]).to.have.property('user-email')
        expect(body.data[0]['user-email']).to.contain('email format')
      })

      const tests = [
        { args: 'key', invalidParam: '', reason: 'empty' },
        { args: 'key', invalidParam: 'a'.repeat(256), reason: 'length must equal or less than' },
        { args: 'bundleId', invalidParam: '', reason: 'empty' },
        { args: 'bundleId', invalidParam: 'a.s', reason: 'bad format.' },
      ]

      tests.forEach((test) => {
        it(`invalid ${test.args}`, async function () {
          app[test.args] = test.invalidParam
          const res = await this.request.post('/apps').send(app).set('user-email', userEmail)
          expect(res.status).to.equal(422)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.data).to.exist()
          expect(body.data).to.have.length(1)
          expect(body.data[0]).to.have.property(test.args)
          expect(body.data[0][test.args]).to.contain(test.reason)
        })
      })
    })
  })

  describe('App Handler', () => {
    let app
    let existingApp

    beforeEach(async function () {
      app = {
        key: uuid.v4(),
        bundleId: `com.app.${uuid.v4().split('-')[0]}`,
      }
      existingApp = await this.app.db.App.create({
        key: uuid.v4(),
        bundleId: `com.app.${uuid.v4().split('-')[0]}`,
        createdBy: 'another@somewhere.com',
      })
    })

    describe('GET', () => {
      it('should return 200 if the app exists', async function () {
        const res = await this.request.get(`/apps/${existingApp.id}`)
        expect(res.status).to.equal(200)

        const body = res.body
        expect(body).to.be.an('object')

        expect(body.app).to.exist()
        expect(body.app).to.be.an('object')

        expect(body.app.key).to.equal(existingApp.key)
        expect(body.app.bundleId).to.equal(existingApp.bundleId)
        expect(body.app.createdBy).to.equal(existingApp.createdBy)
      })

      it('should return 404 if the app does not exist', async function () {
        const res = await this.request.get(`/apps/${uuid.v4()}`)
        expect(res.status).to.equal(404)
      })
    })

    describe('PUT', () => {
      it('should return 200 and the updated app', async function () {
        const res = await this.request.put(`/apps/${existingApp.id}`).send(app)
        expect(res.status).to.equal(200)

        const body = res.body
        expect(body).to.be.an('object')

        expect(body.app).to.exist()
        expect(body.app).to.be.an('object')

        expect(body.app.id).to.exist()
        expect(body.app.key).to.equal(app.key)
        expect(body.app.bundleId).to.equal(app.bundleId)
        expect(body.app.createdBy).to.equal(existingApp.createdBy)

        const dbApp = await this.app.db.App.findById(body.app.id)
        expect(dbApp.key).to.equal(app.key)
        expect(dbApp.bundleId).to.equal(app.bundleId)
        expect(dbApp.createdBy).to.equal(existingApp.createdBy)
      })

      describe('Should fail if', () => {
        it('app does not exist', async function () {
          const res = await this.request.put(`/apps/${uuid.v4()}`).send(app)
          expect(res.status).to.equal(404)
        })

        const tests = [
          { args: 'key' },
          { args: 'bundleId' },
        ]

        tests.forEach((test) => {
          it(`missing ${test.args}`, async function () {
            delete app[test.args]
            const res = await this.request.put(`/apps/${existingApp.id}`).send(app)
            expect(res.status).to.equal(422)

            const body = res.body
            expect(body).to.be.an('object')

            expect(body.data).to.exist()
            expect(body.data).to.have.length(1)
            expect(body.data[0]).to.have.property(test.args)
            expect(body.data[0][test.args]).to.contain('empty')
          })
        })
      })

      const tests = [
        { args: 'key', invalidParam: '', reason: 'empty' },
        { args: 'key', invalidParam: 'a'.repeat(256), reason: 'length must equal or less than' },
        { args: 'bundleId', invalidParam: '', reason: 'empty' },
        { args: 'bundleId', invalidParam: 'a.s', reason: 'bad format.' },
      ]

      tests.forEach((test) => {
        it(`invalid ${test.args}`, async function () {
          app[test.args] = test.invalidParam
          const res = await this.request.put(`/apps/${existingApp.id}`).send(app)
          expect(res.status).to.equal(422)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.data).to.exist()
          expect(body.data).to.have.length(1)
          expect(body.data[0]).to.have.property(test.args)
          expect(body.data[0][test.args]).to.contain(test.reason)
        })
      })
    })

    describe('DELETE', () => {
      it('should return 204 if the app exists', async function () {
        const res = await this.request.delete(`/apps/${existingApp.id}`)
        expect(res.status).to.equal(204)

        const dbApp = await this.app.db.App.findById(existingApp.id)
        expect(dbApp).not.to.exist()
      })

      it('should return 404 if the app does not exist', async function () {
        const res = await this.request.delete(`/apps/${uuid.v4()}`)
        expect(res.status).to.equal(404)
      })
    })
  })
})
