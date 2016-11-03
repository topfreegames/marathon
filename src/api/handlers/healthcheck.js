import { check as redisCheck } from '../../extensions/redis'

export default class HealthcheckHandler {
  constructor(app) {
    this.app = app
    this.route = '/healthcheck'
    this.resetServices()
  }

  resetServices() {
    this.services = {
      redis: { up: false },
    }
  }

  hasFailed() {
    return !this.services.redis.up
  }

  async get(ctx) {
    this.services.redis = await redisCheck(this.app.redisClient)
    ctx.body = JSON.stringify(this.services)

    if (this.hasFailed()) {
      ctx.status = 500
    }
  }
}
