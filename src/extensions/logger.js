// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import bunyan from 'bunyan'
import version from '../extensions/version'

export default class Logger {
  constructor(config) {
    this.config = config
    this.logLevel = config.get('app.log.level')
    this.logToStdOut = this.config.get('app.log.logToStdOut')
    this.logToFile = this.config.get('app.log.logToFile')
    this.logFile = this.config.get('app.log.file')
    this.configureLogger()
  }

  getStreams() {
    const streams = []
    if (this.logToStdOut) {
      streams.push({
        stream: process.stdout,
        level: this.logLevel,
      })
    }
    if (this.logToFile) {
      streams.push({
        path: this.logFile,
        level: this.logLevel,
      })
    }
    return streams
  }

  configureLogger() {
    this.logger = bunyan.createLogger({
      name: this.config.get('app.name'),
      src: false,
      streams: this.getStreams(),
      serializers: { err: bunyan.stdSerializers.err },
    }).child({ version }, true)
  }
}
