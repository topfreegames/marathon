// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import { expect, beforeEachFunc } from '../common'
import uuid from 'uuid'

describe('API', () => {
  describe('Handlers', () => {
    let existingApp

    beforeEach(async function () {
      await beforeEachFunc(this)

      existingApp = await this.app.db.App.create({
        key: uuid.v4(),
        bundleId: `com.app.${uuid.v4().split('-')[0]}`,
        createdBy: 'another@somewhere.com',
      })
    })

    describe('Templates Handler', () => {
      describe('GET', () => {
        it('should return 200 and an empty list of templates', async function () {
          await this.app.db.Template.destroy({ truncate: true, cascade: true })
          const res = await this.request.get(`/apps/${existingApp.id}/templates`)
          expect(res.status).to.equal(200)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.templates).to.exist()
          expect(body.templates).to.have.length(0)
        })

        it('should return 200 and a list of templates', async function () {
          const template = {
            name: uuid.v4(),
            defaults: { default: 'value' },
            body: { my: 'body' },
            compiledBody: 'compiled-body-string',
            locale: 'pt',
            createdBy: 'someone@somewhere.com',
            appId: existingApp.id,
          }
          await this.app.db.Template.create(template)
          const res = await this.request.get(`/apps/${existingApp.id}/templates`)
          expect(res.status).to.equal(200)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.templates).to.exist()
          expect(body.templates).to.have.length.at.least(1)

          const myTemplate = body.templates.filter(t => t.key === template.key)[0]
          expect(myTemplate).to.exist()
          expect(myTemplate.id).to.exist()
          expect(myTemplate.name).to.equal(template.name)
          expect(myTemplate.locale).to.equal(template.locale)
          expect(myTemplate.createdBy).to.equal(template.createdBy)
        })
      })

      describe('POST', () => {
        let template
        let userEmail

        beforeEach(() => {
          template = {
            name: uuid.v4(),
            defaults: JSON.stringify({ default: 'value' }),
            body: JSON.stringify({ my: 'body' }),
            locale: 'pt',
            createdBy: 'someone@somewhere.com',
            appId: existingApp.id,
          }
          userEmail = 'someone@somewhere.com'
        })

        it('should return 201 and the created template', async function () {
          const res = await this.request.post(`/apps/${existingApp.id}/templates`).send(template)
          .set('user-email', userEmail)
          expect(res.status).to.equal(201)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.template).to.exist()
          expect(body.template).to.be.an('object')

          expect(body.template.id).to.exist()
          expect(body.template.compiledBody).to.exist()
          expect(body.template.appId).to.equal(existingApp.id)
          expect(body.template.name).to.equal(template.name)
          expect(body.template.locale).to.equal(template.locale)
          expect(body.template.defaults).to.equal(template.defaults)
          expect(body.template.body).to.equal(template.body)
          expect(body.template.createdBy).to.equal(userEmail)

          const dbTemplate = await this.app.db.Template.findById(body.template.id)
          expect(dbTemplate.appId).to.equal(existingApp.id)
          expect(dbTemplate.name).to.equal(template.name)
          expect(dbTemplate.locale).to.equal(template.locale)
          expect(dbTemplate.defaults).to.equal(template.defaults)
          expect(dbTemplate.body).to.equal(template.body)
          expect(dbTemplate.createdBy).to.equal(userEmail)
        })

        it('should return 201 and the created template with default locale', async function () {
          delete template.locale
          const res = await this.request.post(`/apps/${existingApp.id}/templates`).send(template)
          .set('user-email', userEmail)
          expect(res.status).to.equal(201)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.template).to.exist()
          expect(body.template).to.be.an('object')

          expect(body.template.id).to.exist()
          expect(body.template.locale).to.equal('en')

          const dbTemplate = await this.app.db.Template.findById(body.template.id)
          expect(dbTemplate.locale).to.equal('en')
        })

        describe('Should fail if', () => {
          it('template with same appId, name and locale already exists', async function () {
            await this.request.post(`/apps/${existingApp.id}/templates`).send(template)
            .set('user-email', userEmail)

            const res = await this.request.post(`/apps/${existingApp.id}/templates`).send(template)
            .set('user-email', userEmail)
            expect(res.status).to.equal(409)
          })

          it('app with given appId does not exist', async function () {
            const res = await this.request.post(`/apps/${uuid.v4()}/templates`).send(template)
            .set('user-email', userEmail)
            expect(res.status).to.equal(422)
          })

          it('missing user-email header', async function () {
            const res = await this.request.post(`/apps/${existingApp.id}/templates`).send(template)
            expect(res.status).to.equal(422)

            const body = res.body
            expect(body).to.be.an('object')

            expect(body.data).to.exist()
            expect(body.data).to.have.length(1)
            expect(body.data[0]).to.have.property('user-email')
            expect(body.data[0]['user-email']).to.contain('empty')
          })

          let tests = [
            { args: 'name' },
            { args: 'defaults' },
            { args: 'body' },
          ]

          tests.forEach((test) => {
            it(`missing ${test.args}`, async function () {
              delete template[test.args]
              const res = await this.request.post(`/apps/${existingApp.id}/templates`).send(template)
              .set('user-email', userEmail)
              expect(res.status).to.equal(422)

              const body = res.body
              expect(body).to.be.an('object')

              expect(body.data).to.exist()
              expect(body.data).to.have.length(1)
              expect(body.data[0]).to.have.property(test.args)
              expect(body.data[0][test.args]).to.contain('empty')
            })
          })

          it('invalid user-email header', async function () {
            const res = await this.request.post(`/apps/${existingApp.id}/templates`).send(template)
            .set('user-email', 'not an email')
            expect(res.status).to.equal(422)

            const body = res.body
            expect(body).to.be.an('object')

            expect(body.data).to.exist()
            expect(body.data).to.have.length(1)
            expect(body.data[0]).to.have.property('user-email')
            expect(body.data[0]['user-email']).to.contain('email format')
          })

          tests = [
            { args: 'name', invalidParam: '', reason: 'empty' },
            { args: 'name', invalidParam: 'a'.repeat(256), reason: 'length must equal or less than' },
            { args: 'defaults', invalidParam: 'qwe', reason: 'is not a json format' },
            { args: 'body', invalidParam: 'asd', reason: 'is not a json format' },
            { args: 'locale', invalidParam: '', reason: 'length must equal or great than' },
            { args: 'locale', invalidParam: 'a'.repeat(11), reason: 'length must equal or less than' },
          ]

          tests.forEach((test) => {
            it(`invalid ${test.args}`, async function () {
              template[test.args] = test.invalidParam
              const res = await this.request.post(`/apps/${existingApp.id}/templates`).send(template)
              .set('user-email', userEmail)
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
    })

    describe('Template Handler', () => {
      let template
      let existingTemplate

      beforeEach(async function () {
        template = {
          name: uuid.v4(),
          defaults: JSON.stringify({ default: 'value' }),
          locale: 'pt',
          body: JSON.stringify({ my: 'body' }),
          createdBy: 'someone@somewhere.com',
          appId: existingApp.id,
        }
        existingTemplate = await this.app.db.Template.create({
          name: uuid.v4(),
          locale: 'pt',
          defaults: { default: 'value' },
          body: { my: 'body' },
          compiledBody: 'compiled-body-string',
          createdBy: 'someone@somewhere.com',
          appId: existingApp.id,
        })
      })

      describe('GET', () => {
        it('should return 200 if the template exists for the given app', async function () {
          const res = await this.request.get(
            `/apps/${existingApp.id}/templates/${existingTemplate.id}`)
          expect(res.status).to.equal(200)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.template).to.exist()
          expect(body.template).to.be.an('object')

          expect(body.template.name).to.equal(existingTemplate.name)
          expect(body.template.defaults).to.deep.equal(existingTemplate.defaults)
          expect(body.template.locale).to.equal(existingTemplate.locale)
          expect(body.template.body).to.deep.equal(existingTemplate.body)
          expect(body.template.compiledBody).to.exist()
          expect(body.template.createdBy).to.equal(existingTemplate.createdBy)
          expect(body.template.appId).to.equal(existingTemplate.appId)
          expect(body.template.appId).to.equal(existingApp.id)
        })

        it('should return 404 if the template does not exist', async function () {
          const res = await this.request.get(`/apps/${uuid.v4()}/templates/${existingTemplate.id}`)
          expect(res.status).to.equal(404)
        })

        it('should return 404 if the app does not exist', async function () {
          const res = await this.request.get(`/apps/${existingApp.id}/templates/${uuid.v4()}`)
          expect(res.status).to.equal(404)
        })
      })

      describe('PUT', () => {
        it('should return 200 and the updated template', async function () {
          const res = await this.request
          .put(`/apps/${existingApp.id}/templates/${existingTemplate.id}`).send(template)
          expect(res.status).to.equal(200)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.template).to.exist()
          expect(body.template).to.be.an('object')

          expect(body.template.id).to.exist()
          expect(body.template.name).to.equal(template.name)
          expect(body.template.defaults).to.equal(template.defaults)
          expect(body.template.locale).to.equal(template.locale)
          expect(body.template.body).to.equal(template.body)
          expect(body.template.compiledBody).to.not.equal(template.compiledBody)
          expect(body.template.createdBy).to.equal(existingTemplate.createdBy)
          expect(body.template.appId).to.equal(existingTemplate.appId)
          expect(body.template.appId).to.equal(existingApp.id)

          const dbTemplate = await this.app.db.Template.findById(body.template.id)
          expect(dbTemplate.name).to.equal(template.name)
          expect(dbTemplate.defaults).to.equal(template.defaults)
          expect(dbTemplate.body).to.equal(template.body)
          expect(dbTemplate.compiledBody).to.not.equal(template.body)
          expect(dbTemplate.locale).to.equal(template.locale)
          expect(dbTemplate.createdBy).to.equal(existingTemplate.createdBy)
          expect(dbTemplate.appId).to.equal(existingApp.id)
        })

        describe('Should fail if', () => {
          it('template does not exist', async function () {
            const res = await this.request
            .put(`/apps/${existingApp.id}/templates/${uuid.v4()}`).send(template)
            expect(res.status).to.equal(404)
          })

          it('app does not exist', async function () {
            const res = await this.request
            .put(`/apps/${uuid.v4()}/templates/${existingTemplate.id}`).send(template)
            expect(res.status).to.equal(404)
          })

          let tests = [
            { args: 'name' },
            { args: 'defaults' },
            { args: 'body' },
          ]

          tests.forEach((test) => {
            it(`missing ${test.args}`, async function () {
              delete template[test.args]
              const res = await this.request
              .put(`/apps/${existingApp.id}/templates/${existingTemplate.id}`).send(template)
              expect(res.status).to.equal(422)

              const body = res.body
              expect(body).to.be.an('object')

              expect(body.data).to.exist()
              expect(body.data).to.have.length(1)
              expect(body.data[0]).to.have.property(test.args)
              expect(body.data[0][test.args]).to.contain('empty')
            })
          })

          tests = [
            { args: 'name', invalidParam: '', reason: 'empty' },
            { args: 'name', invalidParam: 'a'.repeat(256), reason: 'length must equal or less than' },
            { args: 'defaults', invalidParam: 'qwe', reason: 'is not a json format' },
            { args: 'body', invalidParam: 'asd', reason: 'is not a json format' },
            { args: 'locale', invalidParam: 'a'.repeat(11), reason: 'length must equal or less than' },
          ]

          tests.forEach((test) => {
            it(`invalid ${test.args}`, async function () {
              template[test.args] = test.invalidParam
              const res = await this.request
              .put(`/apps/${existingApp.id}/templates/${existingTemplate.id}`).send(template)
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

      describe('DELETE', () => {
        it('should return 204 if the app exists', async function () {
          const res = await this.request
            .delete(`/apps/${existingApp.id}/templates/${existingTemplate.id}`)
          expect(res.status).to.equal(204)

          const dbApp = await this.app.db.Template.findById(existingTemplate.id)
          expect(dbApp).not.to.exist()
        })

        it('should return 404 if the template does not exist', async function () {
          const res = await this.request
          .delete(`/apps/${existingApp.id}/templates/${uuid.v4()}`)
          expect(res.status).to.equal(404)
        })

        it('should return 404 if the app does not exist', async function () {
          const res = await this.request
            .delete(`/apps/${uuid.v4()}/templates/${existingTemplate.id}`)
          expect(res.status).to.equal(404)
        })
      })
    })
  })
})
