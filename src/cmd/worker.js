// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import config from 'config'
import MarathonWorkerApp from '../worker/app'

export default class WorkerCmd {
  constructor(rootCmd) {
    rootCmd
      .command('worker')
      .description('Starts Marathon\'s Worker.')
      .action(async () => {
        const app = new MarathonWorkerApp(config)
        await app.initializeWorker()
        app.run()
      })
  }
}
