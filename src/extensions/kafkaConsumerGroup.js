// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import kafka from 'kafka-node'

export async function check(consumerGroup) {
  const result = {
    up: false,
    error: null,
  }

  try {
    result.up = consumerGroup.ready
  } catch (error) {
    result.error = error.message
  }

  return result
}

export async function connect(url, options, logger) {
  const logr = logger.child({
    source: 'kafka-consumer-group-extension',
    url,
    options,
  })
  logr.debug('Connecting to kafka consumerGroup...')

  const opt = Object.assign({}, options, {
    host: url,
  })
  const consumerGroup = new kafka.ConsumerGroup(opt)

  const hasConnected = new Promise((resolve, reject) => {
    if (consumerGroup.ready) {
      logr.debug('Connection to Kafka consumer group has been established successfully.')
      resolve(consumerGroup)
      return
    }

    consumerGroup.on('ready', () => {
      logr.debug('Connection to Kafka consumer group has been established successfully.')
      resolve(consumerGroup)
    })
    consumerGroup.on('error', (err) => {
      logr.error({ err }, 'Connection to Kafka consumer group failed.')
      reject(err)
    })
  })

  await hasConnected

  logr.info('Successfully connected to Kafka consumer group.')
  return consumerGroup
}
