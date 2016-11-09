// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import Boom from 'boom'

import produceJob from '../libs/appProducer'

export class JobsHandler {
  constructor(app) {
    this.app = app
    this.route = '/apps/:id/templates/:tid/jobs'
    this.logger = this.app.logger.child({
      source: 'JobsHandler',
    })
  }

  // async validatePost(ctx, next) {
  //   const logr = this.logger.child({
  //     operation: 'validatePost',
  //   })
  //   ctx.checkHeader('user-email').notEmpty().isEmail()
  //   ctx.checkBody('name').notEmpty().len(1, 255)
  //   ctx.checkBody('defaults').notEmpty().isJSON()
  //   ctx.checkBody('body').notEmpty().isJSON()
  //   ctx.checkBody('locale').optional().len(1, 10)
  //   if (ctx.errors) {
  //     const err = Boom.badData('wrong arguments', ctx.errors)
  //     ctx.status = 422
  //     ctx.body = err.output.payload
  //     ctx.body.data = err.data
  //     logr.warn({ err }, 'Failed validation.')
  //     return
  //   }
  //   await next()
  // }

  async post(ctx) {
    const logr = this.logger.child({
      operation: 'post',
    })
    const body = ctx.request.body
    body.createdBy = ctx.request.header['user-email']
    body.appId = ctx.params.id
    body.templateId = ctx.params.tid
    try {
      const job = await this.app.db.Job.create(body)
      const app = await this.app.App.findById(body.appId)
      const jobPayload = {
        jobId: job.id,
        context: job.context,
        bundleId: app.bundleId,
        service: job.service,
        expiration: job.expireAt,
      }
      await produceJob(this.app.producer, jobPayload, this.logger)
      ctx.body = { job }
      ctx.status = 201
    } catch (err) {
      if (err.name === 'SequelizeForeignKeyConstraintError') {
        const error = Boom.badData('wrong arguments', [err])
        ctx.status = 422
        ctx.body = error.output.payload
        ctx.body.data = error.data
        logr.warn({ err }, 'App or template with given id does not exist.')
        ctx.status = 422
        return
      }
      throw err
    }
  }
}
