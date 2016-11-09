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

  async validatePost(ctx, next) {
    const logr = this.logger.child({
      operation: 'validatePost',
    })
    ctx.checkHeader('user-email').notEmpty().isEmail()
    ctx.checkBody('context').notEmpty().isJSON()
    ctx.checkBody('service').notEmpty().in(['apns', 'gcm'])
    ctx.checkBody('filters').optional().isJSON()
    ctx.checkBody('csvUrl').optional().isUrl()
    ctx.checkBody('expireAt').optional().isDate().toDate()

    if (ctx.request.body.filters && ctx.request.body.csvUrl) {
      const err = [
        { filters: 'filters or csvUrl must exist, not both.' },
        { csvUrl: 'filters or csvUrl must exist, not both.' },
      ]
      if (ctx.errors) {
        ctx.errors = ctx.errors.concat(err)
      } else {
        ctx.errors = err
      }
    }

    if (!ctx.request.body.filters && !ctx.request.body.csvUrl) {
      const err = [
        { filters: 'filters or csvUrl must exist.' },
        { csvUrl: 'filters or csvUrl must exist.' },
      ]
      if (ctx.errors) {
        ctx.errors = ctx.errors.concat(err)
      } else {
        ctx.errors = err
      }
    }

    if (ctx.errors) {
      const err = Boom.badData('wrong arguments', ctx.errors)
      ctx.status = 422
      ctx.body = err.output.payload
      ctx.body.data = err.data
      logr.warn({ err }, 'Failed validation.')
      return
    }
    await next()
  }

  async get(ctx) {
    const jobs = await this.app.db.Job.findAll({
      where: { appId: ctx.params.id, templateId: ctx.params.tid },
    })
    ctx.body = { jobs }
    ctx.status = 200
  }

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
      const app = await this.app.db.App.findById(body.appId)
      const jobPayload = {
        jobId: job.id,
        context: job.context,
        bundleId: app.bundleId,
        service: job.service,
        expiration: job.expireAt,
      }
      await produceJob(this.app.kue, jobPayload, this.logger)
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

export class JobHandler {
  constructor(app) {
    this.app = app
    this.route = '/apps/:id/templates/:tid/jobs/:jid'
    this.logger = this.app.logger.child({
      source: 'JobHandler',
    })
  }

  async get(ctx) {
    const job = await this.app.db.Job.find({
      where: { id: ctx.params.jid, appId: ctx.params.id, templateId: ctx.params.tid },
    })
    if (!job) {
      ctx.status = 404
      return
    }
    ctx.body = { job }
    ctx.status = 200
  }

  // TODO: add stop job
}
