// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

export default async function produceJob(producer, job) {
  const message = {
    jobId: job.id,
    context: job.context,
    app: job.bundleId,
    service: job.service,
    expiration: job.expiration,
  }

  const payload = {
    topic: 'marathonjobs',
    messages: JSON.stringify(message),
  }

  await new Promise((resolve, reject) => {
    producer.send(payload, (err, data) => {
      if (err) {
        reject(err)
        return
      }
      resolve(data)
    })
  })
}
