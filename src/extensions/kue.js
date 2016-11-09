// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import kue from 'kue'

export async function connect(redisClient, options, logger) {
  const logr = logger.child({
    options,
    source: 'kue-extension',
  })

  options.redis = options.redis || {}
  options.redis.createClientFactory = () => redisClient
  options.jobEvents = false

  logr.debug('Creating Kue Queue...')
  const queue = kue.createQueue(options)

  // this is strongly advised by Kue's docs to prevent issues from redis disconnections
  // queue.watchStuckJobs()

  logr.info('Successfully created Kue Queue.')
  return queue
}


export async function createJob(client, queue, data, logger) {
  const logr = logger.child({
    queue,
    data,
  })

  logr.debug('Creating new job...')
  const hasSaved = new Promise((resolve, reject) => {
    const job = client.create(queue, data).removeOnComplete(true)
    job.save((err) => {
      if (err) {
        reject(err)
        return
      }
      resolve(job)
    })
  })

  return await hasSaved
}
