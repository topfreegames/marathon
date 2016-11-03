import { Client } from 'kafka-node'

export async function check(kafkaClient) {
  const result = {
    up: kafkaClient.ready,
    error: null,
  }

  return result
}

export async function connect(url, clientId, logger) {
  const logr = logger.child({
    url,
    clientId,
    source: 'kafka-client-extension',
  })
  logr.debug('Connecting to Kafka...')
  const kafkaClient = new Client(url, clientId)
  const hasConnected = new Promise((resolve, reject) => {
    kafkaClient.on('ready', () => {
      logr.debug('Kafka is ready.')
      resolve(kafkaClient)
    })
    kafkaClient.on('error', (err) => {
      logr.error({ err }, 'Failed to connect to kafka.')
      reject(err)
    })
  })

  logr.debug('Waiting for connection...')
  await hasConnected

  logr.info('Successfully connected to kafka.')
  return kafkaClient
}
