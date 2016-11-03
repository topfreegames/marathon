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
