// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import { camelize } from 'humps'
import koaBodyparser from 'koa-bodyparser'
import koaRouter from 'koa-router'
import koaValidate from 'koa-validate'
import path from 'path'
import Koa from 'koa'
import Logger from '../extensions/logger'
import { AppHandler, AppsHandler } from './handlers/app'
import HealthcheckHandler from './handlers/healthcheck'
import { JobHandler, JobsHandler } from './handlers/job'
import { TemplateHandler, TemplatesHandler } from './handlers/template'
import { connect as kueConnect } from '../extensions/kue'
import { connect as redisConnect, disconnect as redisDisconnect } from '../extensions/redis'
import { connect as pgConnect, disconnect as pgDisconnect } from '../extensions/postgresql'
import { connect as kafkaClientConnect, disconnect as kafkaClientDisconnect } from '../extensions/kafkaClient'
import { connect as kafkaProducerConnect } from '../extensions/kafkaProducer'

process.setMaxListeners(60)

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
    this.pgConfig = config.get('app.services.postgresql')
  }

  getHandlers() { //eslint-disable-line
    const handlers = []

    // Include handlers here
    // Be careful as this will influence the order routes are added to the router
    // Correct order:
    // GET /apps/blabla
    // GET /apps/:id
    // Bad order (will always match the first route):
    // GET /apps/:id
    // GET /apps/blabla
    handlers.push(HealthcheckHandler)
    handlers.push(AppsHandler)
    handlers.push(AppHandler)
    handlers.push(TemplatesHandler)
    handlers.push(TemplateHandler)
    handlers.push(JobsHandler)
    handlers.push(JobHandler)

    return handlers
  }

  exit(err) {
    if (process.env.NODE_ENV === 'test') {
      throw err
    }
    this.logger.fatal({ err })
    process.exit(1)
  }

  configureLogger() {
    this.logger = new Logger(this.config).logger.child({
      source: 'app',
    })
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

  async stopRedis() {
    await redisDisconnect(this.redisClient)
    this.redisClient = null
  }

  async configurePostgreSQL() {
    try {
      this.db = await pgConnect(
        this.pgConfig.url,
        this.pgConfig.options,
        this.logger
      )
    } catch (err) {
      this.exit(err)
    }
  }

  async stopPostgreSQL() {
    await pgDisconnect(this.db)
    this.db = null
  }

  async configureKafka() {
    try {
      this.logger.debug('Connecting API Kafka client...')
      const cfg = this.config.get('app.services.kafka.api.client')
      this.apiKafkaClient = await kafkaClientConnect(cfg.url, cfg.clientId, this.logger)

      this.logger.debug('Connecting API Kafka producer...')
      const producerCfg = this.config.get('app.services.kafka.api.producer')
      this.apiKafkaProducer = await kafkaProducerConnect(
        this.apiKafkaClient,
        producerCfg,
        this.logger
      )
    } catch (err) {
      this.exit(err)
    }
  }

  async stopKafka() {
    await kafkaClientDisconnect(this.apiKafkaClient)
    this.apiKafkaClient = null
    this.apiKafkaProducer = null
  }

  async configureKue() {
    try {
      this.logger.debug('Connecting API Kue client...')
      const cfg = this.config.get('app.services.kue')
      this.kue = await kueConnect(this.redisClient, cfg, this.logger)
    } catch (err) {
      this.exit(err)
    }
  }

  async initializeServices() {
    try {
      this.logger.debug('Starting redis configuration...')
      await this.configureRedis()
      this.logger.debug('Starting PostgreSQL configuration...')
      await this.configurePostgreSQL()
      this.logger.debug('Starting Kafka configuration...')
      await this.configureKafka()
      this.logger.debug('Starting Kue configuration...')
      await this.configureKue()
    } catch (err) {
      this.exit(err)
    }
  }

  async stopServices() {
    try {
      this.logger.debug('Stopping redis...')
      await this.stopRedis()
      this.logger.debug('Stopping PostgreSQL...')
      await this.stopPostgreSQL()
      this.logger.debug('Stopping Kafka...')
      await this.stopKafka()
    } catch (err) {
      this.exit(err)
    }
  }


  configureMiddleware() {
    this.koaApp.use(koaBodyparser())
    this.koaApp.use(async (ctx, next) => {
      const start = new Date()
      await next()
      const ms = new Date() - start
      ctx.set('X-Response-Time', `${ms}ms`)
    })
    koaValidate(this.koaApp)
  }

  async initializeApp() {
    const self = this
    await this.initializeServices()
    const router = koaRouter()
    this.handlers.forEach((HandlerClass) => {
      const inst = new HandlerClass(self)
      this.allowedMethods.forEach((methodName) => {
        if (!inst[methodName]) {
          return
        }
        const handlerMethod = inst[methodName]
        const method = router[methodName]
        const args = [inst.route]
        const validateName = camelize(`validate_${methodName}`)
        if (inst[validateName]) {
          args.push(async (ctx, next) => {
            const handlerInstance = new HandlerClass(self)
            await handlerInstance[validateName].bind(handlerInstance)(ctx, next)
          })
        }
        args.push(async (ctx) => {
          const handlerInstance = new HandlerClass(self)
          await handlerMethod.apply(handlerInstance, [ctx])
        })
        method.apply(router, args)
      })
    })
    this.koaApp.use(router.routes())
    this.koaApp.use(router.allowedMethods())
  }

  async run() {
    const PORT = this.config.get('app.port')
    await this.initializeApp()

    this.logger.info(`Listening on port ${PORT}...`)
    this.koaApp.listen(PORT)
  }

  async stop() {
    await this.stopServices()
  }
}
