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
    let existingTemplate
    beforeEach(async function () {
      await beforeEachFunc(this)

      existingApp = await this.app.db.App.create({
        key: uuid.v4(),
        bundleId: `com.app.${uuid.v4().split('-')[0]}`,
        createdBy: 'another@somewhere.com',
      })

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

    describe('Jobs Handler', () => {
      describe('GET', () => {
        it('should return 200 and an empty list of jobs', async function () {
          await this.app.db.Job.destroy({ truncate: true, cascade: true })
          const res = await this.request
            .get(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs`)
          expect(res.status).to.equal(200)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.jobs).to.exist()
          expect(body.jobs).to.have.length(0)
        })

        it('should return 200 and a list of jobs', async function () {
          const job = {
            totalBatches: 10,
            completedBatches: 10,
            completedAt: new Date(),
            expireAt: new Date(),
            context: { body: 'value' },
            service: 'apns',
            filters: { some: 'filters' },
            csvUrl: 'my.url.com',
            createdBy: 'someone@somewhere.com',
            appId: existingApp.id,
            templateId: existingTemplate.id,
          }
          await this.app.db.Job.create(job)
          const res = await this.request
            .get(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs`)
          expect(res.status).to.equal(200)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.jobs).to.exist()
          expect(body.jobs).to.have.length.at.least(1)

          const myJob = body.jobs.filter(t => t.key === job.key)[0]
          expect(myJob).to.exist()
          expect(myJob.id).to.exist()
          expect(myJob.totalBatches).to.equal(job.totalBatches)
          expect(myJob.completedBatches).to.equal(job.completedBatches)
          expect(myJob.completedAt).to.equal(job.completedAt.toISOString())
          expect(myJob.expireAt).to.equal(job.expireAt.toISOString())
          expect(myJob.context).to.deep.equal(job.context)
          expect(myJob.service).to.equal(job.service)
          expect(myJob.filters).to.deep.equal(job.filters)
          expect(myJob.csvUrl).to.equal(job.csvUrl)
          expect(myJob.createdBy).to.equal(job.createdBy)
          expect(myJob.appId).to.equal(existingApp.id)
          expect(myJob.templateId).to.equal(existingTemplate.id)
        })
      })

      describe('POST', () => {
        let job
        let userEmail

        beforeEach(() => {
          job = {
            expireAt: new Date(),
            context: JSON.stringify({ body: 'value' }),
            service: 'apns',
            filters: JSON.stringify({ some: 'filters' }),
            csvUrl: 'my.url.com',
          }
          userEmail = 'someone@somewhere.com'
        })

        it('should return 201 and the created job with filters', async function () {
          delete job.csvUrl
          const res = await this.request
            .post(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs`).send(job)
            .set('user-email', userEmail)
          expect(res.status).to.equal(201)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.job).to.exist()
          expect(body.job).to.be.an('object')

          expect(body.job.id).to.exist()
          expect(body.job.totalBatches).to.not.exist()
          expect(body.job.completedBatches).to.equal(0)
          expect(body.job.completedAt).to.not.exist()
          expect(body.job.expireAt).to.equal(job.expireAt.toISOString())
          expect(body.job.context).to.deep.equal(job.context)
          expect(body.job.service).to.equal(job.service)
          expect(body.job.filters).to.deep.equal(job.filters)
          expect(body.job.csvUrl).to.to.not.exist()
          expect(body.job.createdBy).to.equal(userEmail)
          expect(body.job.appId).to.equal(existingApp.id)
          expect(body.job.templateId).to.equal(existingTemplate.id)

          const dbJob = await this.app.db.Job.findById(body.job.id)
          expect(dbJob.totalBatches).to.not.exist()
          expect(dbJob.completedBatches).to.equal(0)
          expect(dbJob.completedAt).to.not.exist()
          expect(dbJob.expireAt.toISOString()).to.equal(job.expireAt.toISOString())
          expect(dbJob.context).to.deep.equal(job.context)
          expect(dbJob.service).to.equal(job.service)
          expect(dbJob.filters).to.deep.equal(job.filters)
          expect(dbJob.csvUrl).to.to.not.exist()
          expect(dbJob.createdBy).to.equal(userEmail)
          expect(dbJob.appId).to.equal(existingApp.id)
          expect(dbJob.templateId).to.equal(existingTemplate.id)
        })

        it('should return 201 and the created job with csvUrl', async function () {
          delete job.filters
          const res = await this.request
            .post(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs`).send(job)
            .set('user-email', userEmail)
          expect(res.status).to.equal(201)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.job).to.exist()
          expect(body.job).to.be.an('object')

          expect(body.job.id).to.exist()
          expect(body.job.totalBatches).to.not.exist()
          expect(body.job.completedBatches).to.equal(0)
          expect(body.job.completedAt).to.not.exist()
          expect(body.job.expireAt).to.equal(job.expireAt.toISOString())
          expect(body.job.context).to.deep.equal(job.context)
          expect(body.job.service).to.equal(job.service)
          expect(body.job.csvUrl).to.equal(job.csvUrl)
          expect(body.job.filters).to.to.not.exist()
          expect(body.job.createdBy).to.equal(userEmail)
          expect(body.job.appId).to.equal(existingApp.id)
          expect(body.job.templateId).to.equal(existingTemplate.id)

          const dbJob = await this.app.db.Job.findById(body.job.id)
          expect(dbJob.totalBatches).to.not.exist()
          expect(dbJob.completedBatches).to.equal(0)
          expect(dbJob.completedAt).to.not.exist()
          expect(dbJob.expireAt.toISOString()).to.equal(job.expireAt.toISOString())
          expect(dbJob.context).to.deep.equal(job.context)
          expect(dbJob.service).to.equal(job.service)
          expect(dbJob.csvUrl).to.equal(job.csvUrl)
          expect(dbJob.filters).to.to.not.exist()
          expect(dbJob.createdBy).to.equal(userEmail)
          expect(dbJob.appId).to.equal(existingApp.id)
          expect(dbJob.templateId).to.equal(existingTemplate.id)
        })

        it('should return 201 and the created job without expireAt', async function () {
          delete job.csvUrl
          delete job.expireAt
          const res = await this.request
            .post(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs`).send(job)
            .set('user-email', userEmail)
          expect(res.status).to.equal(201)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.job).to.exist()
          expect(body.job).to.be.an('object')

          expect(body.job.id).to.exist()
          expect(body.job.expireAt).to.not.exist()

          const dbJob = await this.app.db.Job.findById(body.job.id)
          expect(dbJob.expireAt).to.not.exist()
        })

        describe('Should fail if', () => {
          it('both filters and csvUrl are not provided', async function () {
            delete job.csvUrl
            delete job.filters
            const res = await this.request
              .post(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs`).send(job)
              .set('user-email', userEmail)
            expect(res.status).to.equal(422)
            expect(res.body.data).to.have.length(2)
          })

          it('both filters and csvUrl are provided', async function () {
            const res = await this.request
              .post(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs`).send(job)
              .set('user-email', userEmail)
            expect(res.status).to.equal(422)
            expect(res.body.data).to.have.length(2)
          })

          it('app with given appId does not exist', async function () {
            delete job.csvUrl
            const res = await this.request
              .post(`/apps/${uuid.v4()}/templates/${existingTemplate.id}/jobs`).send(job)
              .set('user-email', userEmail)
            expect(res.status).to.equal(422)
          })

          it('template with given templateId does not exist', async function () {
            delete job.csvUrl
            const res = await this.request
              .post(`/apps/${existingApp.id}/templates/${uuid.v4()}/jobs`).send(job)
              .set('user-email', userEmail)
            expect(res.status).to.equal(422)
          })

          it('missing user-email header', async function () {
            delete job.csvUrl
            const res = await this.request
              .post(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs`).send(job)
            expect(res.status).to.equal(422)

            const body = res.body
            expect(body).to.be.an('object')

            expect(body.data).to.exist()
            expect(body.data).to.have.length(1)
            expect(body.data[0]).to.have.property('user-email')
            expect(body.data[0]['user-email']).to.contain('empty')
          })

          let tests = [
            { args: 'context' },
            { args: 'service' },
          ]

          tests.forEach((test) => {
            it(`missing ${test.args}`, async function () {
              delete job.csvUrl
              delete job[test.args]
              const res = await this.request
                .post(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs`).send(job)
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
            delete job.csvUrl
            const res = await this.request
              .post(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs`).send(job)
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
            { args: 'context', invalidParam: '', reason: 'empty' },
            { args: 'context', invalidParam: 'qwe', reason: 'is not a json format' },
            { args: 'service', invalidParam: '', reason: 'empty' },
            { args: 'service', invalidParam: 'blabla', reason: 'must be in [apns,gcm]' },
            { args: 'filters', invalidParam: 'asd', reason: 'is not a json format' },
            { args: 'csvUrl', invalidParam: 'blabla', reason: 'is not url format' },
            { args: 'expireAt', invalidParam: 'blabla', reason: 'is not a date format' },
          ]

          tests.forEach((test) => {
            it(`invalid ${test.args}`, async function () {
              if (test.args === 'csvUrl') {
                delete job.filters
              } else {
                delete job.csvUrl
              }
              job[test.args] = test.invalidParam
              const res = await this.request
                .post(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs`).send(job)
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

    describe('Job Handler', () => {
      let existingJob

      beforeEach(async function () {
        existingJob = await this.app.db.Job.create({
          totalBatches: 10,
          completedBatches: 10,
          completedAt: new Date(),
          expireAt: new Date(),
          context: { body: 'value' },
          service: 'apns',
          filters: { some: 'filters' },
          csvUrl: 'my.url.com',
          createdBy: 'someone@somewhere.com',
          appId: existingApp.id,
          templateId: existingTemplate.id,
        })
      })

      describe('GET', () => {
        it('should return 200 if the job exists for the given app and template', async function () {
          const res = await this.request
            .get(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs/${existingJob.id}`)

          expect(res.status).to.equal(200)

          const body = res.body
          expect(body).to.be.an('object')

          expect(body.job).to.exist()
          expect(body.job).to.be.an('object')

          expect(body.job.id).to.equal(existingJob.id)
          expect(body.job.totalBatches).to.equal(existingJob.totalBatches)
          expect(body.job.completedBatches).to.equal(existingJob.completedBatches)
          expect(body.job.completedAt).to.equal(existingJob.completedAt.toISOString())
          expect(body.job.expireAt).to.equal(existingJob.expireAt.toISOString())
          expect(body.job.context).to.deep.equal(existingJob.context)
          expect(body.job.service).to.equal(existingJob.service)
          expect(body.job.filters).to.deep.equal(existingJob.filters)
          expect(body.job.csvUrl).to.equal(existingJob.csvUrl)
          expect(body.job.createdBy).to.equal(existingJob.createdBy)
          expect(body.job.appId).to.equal(existingApp.id)
          expect(body.job.templateId).to.equal(existingTemplate.id)
        })

        it('should return 404 if the job does not exist', async function () {
          const res = await this.request
            .get(`/apps/${existingApp.id}/templates/${existingTemplate.id}/jobs/${uuid.v4()}`)
          expect(res.status).to.equal(404)
        })

        it('should return 404 if the template does not exist', async function () {
          const res = await this.request
            .get(`/apps/${existingApp.id}/templates/${uuid.v4()}/jobs/${existingJob.id}`)
          expect(res.status).to.equal(404)
        })

        it('should return 404 if the app does not exist', async function () {
          const res = await this.request
            .get(`/apps/${uuid.v4()}/templates/${existingTemplate.id}/jobs/${existingJob.id}`)
          expect(res.status).to.equal(404)
        })
      })
    })
  })
})
