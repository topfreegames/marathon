import { check as redisCheck } from '../../extensions/redis'
import { check as pgCheck } from '../../extensions/postgresql'

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
    }
  }

  hasFailed() {
    return !this.services.redis.up || !this.services.postgreSQL.up
  }

  async get(ctx) {
    this.services.redis = await redisCheck(this.app.redisClient)
    this.services.postgreSQL = await pgCheck(this.app.db)
    ctx.body = JSON.stringify(this.services)

    if (this.hasFailed()) {
      ctx.status = 500
    }
  }
}
