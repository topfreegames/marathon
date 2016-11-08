const Boom = require('boom')

export class AppsHandler {
  constructor(app) {
    this.app = app
    this.route = '/apps'
  }

  async validatePost(ctx, next) {
    ctx.checkHeader('user-email').notEmpty().isEmail()
    ctx.checkBody('bundleId').notEmpty().match(/^[a-z0-9]+\.[a-z0-9]+(\.[a-z0-9]+)+$/i)
    ctx.checkBody('key').notEmpty().len(1, 255)
    if (ctx.errors) {
      const error = Boom.badData('wrong arguments', ctx.errors)
      ctx.status = 422
      ctx.body = error.output.payload
      ctx.body.data = error.data
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
    const body = ctx.request.body
    body.createdBy = ctx.request.header['user-email']
    try {
      const app = await this.app.db.App.create(body)
      ctx.body = { app }
      ctx.status = 201
    } catch (err) {
      if (err.name === 'SequelizeUniqueConstraintError') {
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
  }

  async validatePut(ctx, next) {
    ctx.checkBody('bundleId').notEmpty().match(/^[a-z0-9]+\.[a-z0-9]+(\.[a-z0-9]+)+$/i)
    ctx.checkBody('key').notEmpty().len(1, 255)
    if (ctx.errors) {
      const error = Boom.badData('wrong arguments', ctx.errors)
      ctx.status = 422
      ctx.body = error.output.payload
      ctx.body.data = error.data
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
