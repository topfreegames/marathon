// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import kafka from 'kafka-node'

export async function check(producer) {
  const result = {
    up: false,
    error: null,
  }

  try {
    result.up = producer.ready
  } catch (error) {
    result.error = error.message
  }

  return result
}

export async function connect(client, options, logger) {
  const logr = logger.child({
    source: 'kafka-producer-extension',
    options,
  })
  logr.debug('Connecting to kafka producer...')
  const producer = new kafka.Producer(client, options)

  const hasConnected = new Promise((resolve, reject) => {
    if (producer.ready) {
      logr.debug('Connection to Kafka producer has been established successfully.')
      resolve(producer)
      return
    }

    producer.on('ready', () => {
      logr.debug('Connection to Kafka producer has been established successfully.')
      resolve(producer)
    })
    producer.on('error', (err) => {
      logr.error({ err }, 'Connection to Kafka producer failed.')
      reject(err)
    })
  })

  await hasConnected

  logr.info('Successfully connected to Kafka producer.')
  return producer
}
