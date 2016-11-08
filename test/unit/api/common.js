import { expect as exp, chai as chaiMod } from '../../common'
import config from 'config'
import * as sap from 'supertest-as-promised'

import MarathonApp from '../../../src/api/app'

export const expect = exp
export const chai = chaiMod

let PORT = 9000
let app = null

// Before each test create and destroy the app if it does not exist
beforeEach(async function () {
  PORT += 1
  config.app.port = PORT
  if (!app) {
    app = new MarathonApp(config)
  }
  this.app = app
  this.request = sap.agent(this.app.koaApp.listen())
  await this.app.run()
})
