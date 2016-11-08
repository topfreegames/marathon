// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import { expect as exp, chai as chaiMod } from '../../common'
import config from 'config'

import MarathonWorkerApp from '../../../src/worker/app'

export const expect = exp
export const chai = chaiMod

// Before each test create and destroy the app if it does not exist
beforeEach(async function () {
  this.worker = new MarathonWorkerApp(config)
  await this.worker.initializeWorker()
})
