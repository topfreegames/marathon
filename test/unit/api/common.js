// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import { expect as exp, chai as chaiMod } from '../../common'
import config from 'config'
import * as sap from 'supertest-as-promised'

import MarathonApp from '../../../src/api/app'

export const expect = exp
export const chai = chaiMod

let PORT = 9000

export async function beforeEachFunc(self) {
  PORT += 1
  config.app.port = PORT
  const app = new MarathonApp(config)
  self.app = app
  self.request = sap.agent(self.app.koaApp.listen())
  await self.app.run()
}
