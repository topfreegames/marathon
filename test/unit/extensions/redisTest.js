// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import { expect, beforeEachFunc } from './common'
import { check as redisCheck, connect as redisConnect } from '../../../src/extensions/redis'

describe('Extensions', () => {
  describe('Redis Extension', () => {
    beforeEach(async function () {
      await beforeEachFunc(this)
    })

    describe('check', () => {
      it('should check successfully if the connection to redis is up', async function () {
        const redisConfig = this.config.get('app.services.redis')
        const redisClient = await redisConnect(redisConfig.url, { db: redisConfig.db },
            this.logger)
        const result = await redisCheck(redisClient)
        expect(result).to.exist()
        expect(result.up).to.be.true()
      })
    })
    describe('connect', () => {
      it('should connect successfully to redis', async function () {
        const redisConfig = this.config.get('app.services.redis')
        const redisClient = await redisConnect(
          redisConfig.url, { db: redisConfig.db },
          this.logger
        )
        expect(redisClient).to.exist()
      })

      it('should throw error if redis is not up', async function () {
        const redisConfig = this.config.get('app.services.redis')
        const faultyUrl = '//localhost:4334'
        try {
          await redisConnect(faultyUrl, { db: redisConfig.db }, this.logger)
        } catch (e) {
          expect(e).to.exist()
          return
        }
        expect(false).to.be.ok()
      })
    })
  })
})
