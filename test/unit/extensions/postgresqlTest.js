// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import Sequelize from 'sequelize'
import { expect, beforeEachFunc } from './common'
import { check as pgCheck, connect as pgConnect } from '../../../src/extensions/postgresql'

describe('Extensions', () => {
  describe('PostgreSQL Extension', () => {
    beforeEach(async function () {
      await beforeEachFunc(this)
    })

    describe('check', () => {
      it('should check successfully if the connection to postgresql is up', async function () {
        const pgConfig = this.config.get('app.services.postgresql')
        const pgClient = await pgConnect(pgConfig.url, { db: pgConfig.db },
            this.logger)
        const result = await pgCheck(pgClient)
        expect(result).to.exist()
        expect(result.up).to.be.true()
      })

      it('should fail check if the connection to postgresql is wrong', async () => {
        const client = new Sequelize('postgres://marathonfake@localhost:22222/marathonfake', {})
        const result = await pgCheck({ client })
        expect(result).to.exist()
        expect(result.up).to.be.false()
        expect(result.error).to.equal('role "marathonfake" does not exist')
      })
    })

    describe('connect', () => {
      it('should connect successfully to pg', async function () {
        const pgConfig = this.config.get('app.services.postgresql')
        const pgClient = await pgConnect(pgConfig.url, { db: pgConfig.db },
            this.logger)
        expect(pgClient).to.exist()
      })

      it('should throw error if pg is not up', async function () {
        const pgConfig = this.config.get('app.services.postgresql')
        const faultyUrl = '//localhost:4334'
        try {
          await pgConnect(faultyUrl, { db: pgConfig.db }, this.logger)
        } catch (e) {
          expect(e).to.exist()
          return
        }
        expect(false).to.be.ok()
      })
    })
  })
})
