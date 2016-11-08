// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import { expect } from '../common'

describe('Handlers', () => {
  describe('Healthcheck Handler', () => {
    it('should return 200 if all services up', async function () {
      const res = await this.request.get('/healthcheck')
      expect(res.status).to.equal(200)

      const body = res.text
      expect(body).not.to.equal('')

      const result = JSON.parse(body)

      expect(result.redis).to.exist()
      expect(result.redis.up).to.equal(true)
    })

    it('should fail if redis is not up', async function () {
      this.app.redisClient.end(true)

      const res = await this.request.get('/healthcheck')
      expect(res.status).to.equal(500)
    })
  })
})
