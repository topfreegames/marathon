// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

const Boom = require('boom')

export class AppsHandler {
  constructor(app) {
    this.app = app
    this.route = '/apps'
    this.logger = this.app.logger.child({
      source: 'AppsHandler',
    })
  }

  async validatePost(ctx, next) {
    const logr = this.logger.child({
      operation: 'validatePost',
    })
    ctx.checkHeader('user-email').notEmpty().isEmail()
    ctx.checkBody('bundleId').notEmpty().match(/^[a-z0-9]+\.[a-z0-9]+(\.[a-z0-9]+)+$/i)
    ctx.checkBody('key').notEmpty().len(1, 255)
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
    const apps = await this.app.db.App.findAll()
    ctx.body = { apps }
    ctx.status = 200
  }

  async post(ctx) {
    const logr = this.logger.child({
      operation: 'post',
    })
    const body = ctx.request.body
    body.createdBy = ctx.request.header['user-email']
    try {
      const app = await this.app.db.App.create(body)
      ctx.body = { app }
      ctx.status = 201
    } catch (err) {
      if (err.name === 'SequelizeUniqueConstraintError') {
        logr.warn({ err }, 'App with same bundle already exists.')
        ctx.status = 409
        return
      }
      throw err
    }
  }
}

export class AppHandler {
  constructor(app) {
    this.app = app
    this.route = '/apps/:id'
    this.logger = this.app.logger.child({
      source: 'AppHandler',
    })
  }

  async validatePut(ctx, next) {
    const logr = this.logger.child({
      operation: 'validatePut',
    })
    ctx.checkBody('bundleId').notEmpty().match(/^[a-z0-9]+\.[a-z0-9]+(\.[a-z0-9]+)+$/i)
    ctx.checkBody('key').notEmpty().len(1, 255)
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
    const app = await this.app.db.App.findById(ctx.params.id)
    if (!app) {
      ctx.status = 404
      return
    }
    ctx.body = { app }
    ctx.status = 200
  }

  async put(ctx) {
    const body = ctx.request.body
    const app = await this.app.db.App.findById(ctx.params.id)
    if (!app) {
      ctx.status = 404
      return
    }

    const updatedApp = await app.updateAttributes(body)
    ctx.body = { app: updatedApp }
    ctx.status = 200
  }

  async delete(ctx) {
    const app = await this.app.db.App.findById(ctx.params.id)
    if (!app) {
      ctx.status = 404
      return
    }

    await app.destroy()
    ctx.status = 204
  }
}
