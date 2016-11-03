import path from 'path'
import fs from 'fs'
import Koa from 'koa'
import _ from 'koa-route'
import Logger from '../extensions/logger'
import { connect as redisConnect } from '../extensions/redis'


export default class MarathonApp {
  constructor(config) {
    this.config = config
    this.allowedMethods = ['get', 'post', 'put', 'delete']
    this.koaApp = new Koa()
    this.configureLogger()
    this.configureMiddleware()

    this.handlersPath = path.join(__dirname, '../api/handlers')
    this.handlers = this.getHandlers()
    this.redisConfig = config.get('app.services.redis')
  }

  getHandlers() {
    const self = this
    const handlers = []
    fs
      .readdirSync(this.handlersPath)
      .forEach((file) => {
        const Handler = require(`./handlers/${file}`).default  // eslint-disable-line
        handlers.push(new Handler(self))
      })

    return handlers
  }

  exit(err) {
    if (process.env.NODE_ENV === 'test') {
      return
    }
    this.logger.fatal({ err })
    process.exit(1)
  }

  configureLogger() {
    this.logger = new Logger(this.config).logger
  }

  async configureRedis() {
    const redisOptions = {
      db: this.redisConfig.db,
      shouldReconnect: this.redisConfig.shouldReconnect,
      password: this.redisConfig.password,
    }
    if (!redisOptions.password) delete redisOptions.password
    try {
      this.redisClient = await redisConnect(
        this.redisConfig.url,
        redisOptions,
        this.logger
      )
    } catch (err) {
      this.exit(err)
    }
  }

  async initializeServices() {
    try {
      await this.configureRedis()
    } catch (err) {
      this.exit(err)
    }
  }

  configureMiddleware() {
    this.koaApp.use(async (ctx, next) => {
      const start = new Date()
      await next()
      const ms = new Date() - start
      ctx.set('X-Response-Time', `${ms}ms`)
    })
  }

  async initializeApp() {
    await this.initializeServices()
    this.handlers.forEach((handler) => {
      this.allowedMethods.forEach((methodName) => {
        if (!handler[methodName]) {
          return
        }
        const handlerMethod = handler[methodName].bind(handler)
        const method = _[methodName]
        this.koaApp.use(
          method(handler.route, async (ctx) => {
            await handlerMethod(ctx)
          })
        )
      })
    })
  }

  async run() {
    const PORT = this.config.get('app.port')
    await this.initializeApp()
    this.koaApp.listen(PORT)
  }
}
