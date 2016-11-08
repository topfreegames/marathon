// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import config from 'config'
import MarathonApp from '../api/app'

export default class StartCmd {
  constructor(rootCmd) {
    rootCmd
      .command('start')
      .description('Start API')
      .action(() => {
        const app = new MarathonApp(config)
        app.run()
      })
  }
}
