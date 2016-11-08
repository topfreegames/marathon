import Logger from '../extensions/logger'
import { connect as redisConnect } from '../extensions/redis'
import { connect as pgConnect } from '../extensions/postgresql'
import { connect as kafkaClientConnect, disconnect as kafkaClientDisconnect } from '../extensions/kafkaClient'
import { connect as kafkaProducerConnect } from '../extensions/kafkaProducer'

const timeout = ms => new Promise(resolve => setTimeout(resolve, ms))

export default class MarathonWorkerApp {
  constructor(config) {
    this.config = config
    this.configureLogger()

    this.redisConfig = config.get('app.services.redis')
    this.pgConfig = config.get('app.services.postgresql')
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
      source: 'worker',
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
    await kafkaClientDisconnect()
    this.apiKafkaClient = null
    this.apiKafkaProducer = null
  }

  async initializeServices() {
    try {
      this.logger.debug('Starting redis configuration...')
      await this.configureRedis()
      this.logger.debug('Starting PostgreSQL configuration...')
      await this.configurePostgreSQL()
      this.logger.debug('Starting Kafka configuration...')
      await this.configureKafka()
    } catch (err) {
      this.exit(err)
    }
  }

  async stopServices() {
    try {
      // this.logger.debug('Starting redis configuration...')
      // await this.configureRedis()
      // this.logger.debug('Starting PostgreSQL configuration...')
      // await this.configurePostgreSQL()
      this.logger.debug('Starting Kafka configuration...')
      await this.stopKafka()
    } catch (err) {
      this.exit(err)
    }
  }

  async initializeWorker() {
    await this.initializeServices()
  }

  getNextJobBatch() {
    return null
  }

  processJobBatch(job) {
    return false
  }

  async iteration() {
    const logr = this.logger.child({
      operation: 'iteration',
    })
    logr.debug('Processing next job...')
    const job = await this.getNextJobBatch()
    if (!job) {
      logr.debug('No jobs found.')
      return false
    }

    logr.debug({ job }, 'Starting processing of job...')
    return await this.processJobBatch(job)
  }

  async run() {
    this.shouldRun = true
    const timeoutInMs = this.config.get('app.worker.loopTimeout')
    this.logger.info('Listening for new job batches...')

    while (this.shouldRun) {
      const res = await this.iteration()

      if (!res) await timeout(timeoutInMs)
    }
  }

  async stop() {
    await this.stopServices()
  }
}
