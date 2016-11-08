// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import { check as redisCheck } from '../../extensions/redis'
import { check as pgCheck } from '../../extensions/postgresql'
import { check as kafkaClientCheck } from '../../extensions/kafkaClient'
import { check as kafkaProducerCheck } from '../../extensions/kafkaProducer'

export default class HealthcheckHandler {
  constructor(app) {
    this.app = app
    this.route = '/healthcheck'
    this.resetServices()
  }

  resetServices() {
    this.services = {
      redis: { up: false },
      postgreSQL: { up: false },
      apiKafkaClient: { up: false },
      apiKafkaProducer: { up: false },
    }
  }

  hasFailed() {
    return (
      !this.services.redis.up ||
      !this.services.postgreSQL.up ||
      !this.services.apiKafkaClient.up ||
      !this.services.apiKafkaProducer.up
    )
  }

  async get(ctx) {
    this.services.redis = await redisCheck(this.app.redisClient)
    this.services.postgreSQL = await pgCheck(this.app.db)
    this.services.apiKafkaClient = await kafkaClientCheck(this.app.apiKafkaClient)
    this.services.apiKafkaProducer = await kafkaProducerCheck(this.app.apiKafkaProducer)
    ctx.body = JSON.stringify(this.services)

    if (this.hasFailed()) {
      ctx.status = 500
    }
  }
}
