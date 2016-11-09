// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import { createJob } from '../../extensions/kue'

export default async function produceJob(client, job, logger) {
  const res = await createJob(
    client,
    'marathon-jobs',
    job,
    logger,
  )
  return res
}
