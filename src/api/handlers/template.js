// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import dust from 'dustjs-linkedin'
import Boom from 'boom'

export class TemplatesHandler {
  constructor(app) {
    this.app = app
    this.route = '/apps/:id/templates'
    this.logger = this.app.logger.child({
      source: 'TemplatesHandler',
    })
  }

  async validatePost(ctx, next) {
    const logr = this.logger.child({
      operation: 'validatePost',
    })
    ctx.checkHeader('user-email').notEmpty().isEmail()
    ctx.checkBody('name').notEmpty().len(1, 255)
    ctx.checkBody('defaults').notEmpty().isJSON()
    ctx.checkBody('body').notEmpty().isJSON()
    ctx.checkBody('locale').optional().len(1, 10)
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
    const templates = await this.app.db.Template.findAll({ where: { appId: ctx.params.id } })
    ctx.body = {
      templates: templates.map(t => ({
        id: t.id,
        name: t.name,
        locale: t.locale,
        appId: t.appId,
        createdBy: t.createdBy,
      })),
    }
    ctx.status = 200
  }

  async post(ctx) {
    const logr = this.logger.child({
      operation: 'post',
    })
    const body = ctx.request.body
    body.createdBy = ctx.request.header['user-email']
    body.appId = ctx.params.id
    body.compiledBody = dust.compile(body.body, 'body')
    try {
      const template = await this.app.db.Template.create(body)
      ctx.body = { template }
      ctx.status = 201
    } catch (err) {
      if (err.name === 'SequelizeUniqueConstraintError') {
        const error = Boom.conflict('conflict', [err])
        ctx.status = 409
        ctx.body = error.output.payload
        ctx.body.data = error.data
        logr.warn({ err }, 'Template with same name and appId already exists.')
        return
      }
      if (err.name === 'SequelizeForeignKeyConstraintError') {
        const error = Boom.badData('wrong arguments', [err])
        ctx.status = 422
        ctx.body = error.output.payload
        ctx.body.data = error.data
        logr.warn({ err }, 'App with given appId does not exist.')
        ctx.status = 422
        return
      }
      throw err
    }
  }
}

export class TemplateHandler {
  constructor(app) {
    this.app = app
    this.route = '/apps/:id/templates/:tid'
    this.logger = this.app.logger.child({
      source: 'TemplateHandler',
    })
  }

  async validatePut(ctx, next) {
    const logr = this.logger.child({
      operation: 'validatePut',
    })
    ctx.checkBody('name').notEmpty().len(1, 255)
    ctx.checkBody('defaults').notEmpty().isJSON()
    ctx.checkBody('body').notEmpty().isJSON()
    ctx.checkBody('locale').optional().len(1, 10)
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
    const template = await this.app.db.Template.find({
      where: { id: ctx.params.tid, appId: ctx.params.id },
    })
    if (!template) {
      ctx.status = 404
      return
    }
    ctx.body = { template }
    ctx.status = 200
  }

  async put(ctx) {
    const body = ctx.request.body
    const template = await this.app.db.Template.find({
      where: { id: ctx.params.tid, appId: ctx.params.id },
    })
    if (!template) {
      ctx.status = 404
      return
    }

    body.compiledBody = dust.compile(body.body, 'body')
    const updatedTemplate = await template.updateAttributes(body)
    ctx.body = { template: updatedTemplate }
    ctx.status = 200
  }

  async delete(ctx) {
    const template = await this.app.db.Template.find({
      where: { id: ctx.params.tid, appId: ctx.params.id },
    })
    if (!template) {
      ctx.status = 404
      return
    }

    await template.destroy()
    ctx.status = 204
  }
}
