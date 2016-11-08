// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import program from 'commander'
import StartCmd from './cmd/start'
import WorkerCmd from './cmd/worker'
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
  new WorkerCmd(cmd.rootCmd) //eslint-disable-line
  cmd.run(process.argv)
}
