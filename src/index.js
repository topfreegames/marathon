import program from 'commander'
import StartCmd from './cmd/start'
import { version } from './extensions/version'

export default class RootCmd {
  constructor() {
    this.rootCmd = program
      .version(version)
  }

  run(args) {
    this.rootCmd.parse(args)
  }
}

if (!module.parent) {
  const cmd = new RootCmd()
  new StartCmd(cmd.rootCmd) //eslint-disable-line
  cmd.run(process.argv)
}
