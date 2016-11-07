import { expect } from '../common'
import { check as redisCheck, connect as redisConnect } from '../../../src/extensions/redis'

describe('Extensions', () => {
  describe('Redis Extension', () => {
    describe('check', () => {
      it('should check successfully if the connection to redis is up', async function () {
        const redisConfig = this.app.config.get('app.services.redis')
        const redisClient = await redisConnect(redisConfig.url, { db: redisConfig.db },
            this.app.logger)
        const result = await redisCheck(redisClient)
        expect(result).to.exist()
        expect(result.up).to.be.true()
      })
    })
    describe('connect', () => {
      it('should connect successfully to redis', async function () {
        const redisConfig = this.app.config.get('app.services.redis')
        const redisClient = await redisConnect(
          redisConfig.url, { db: redisConfig.db },
          this.app.logger
        )
        expect(redisClient).to.exist()
      })

      it('should throw error if redis is not up', async function () {
        const redisConfig = this.app.config.get('app.services.redis')
        const faultyUrl = '//localhost:4334'
        try {
          await redisConnect(faultyUrl, { db: redisConfig.db }, this.app.logger)
        } catch (e) {
          expect(e).to.exist()
          return
        }
        expect(false).to.be.ok()
      })
    })
  })
})
