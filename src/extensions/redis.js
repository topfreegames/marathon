// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import redis from 'redis'
import bluebird from 'bluebird'
import parser from 'redis-info'
import Redlock from 'redlock'

bluebird.promisifyAll(redis.RedisClient.prototype)
bluebird.promisifyAll(redis.Multi.prototype)

let Lock = null
export const LockKey = 'lock::matchmaking'
export const LockTTL = 100

export async function check(redisClient) {
  const result = {
    up: false,
    error: null,
    uptime: null,
    connectedClients: null,
    blockedClients: null,
    usedMemory: null,
    totalSystemMemory: null,
    maxMemory: null,
    rejectedConnections: 0,
    cpuUsage: 0,
  }

  try {
    const res = await redisClient.infoAsync()
    if (res) {
      const info = parser.parse(res)
      result.up = true
      result.uptime = info.uptime_in_seconds
      result.connectedClients = info.connected_clients
      result.blockedClients = info.blocked_clients
      result.usedMemory = info.used_memory_human
      result.totalSystemMemory = info.total_system_memory_human
      result.maxMemory = info.maxmemory_human
      result.rejectedConnections = info.rejected_connections
      result.cpuUsage = info.used_cpu_user
    } else {
      result.error = 'Could not get server status!'
    }
  } catch (error) {
    result.error = error.message
  }

  return result
}

export async function connect(redisUrl, options, logger) {
  const logr = logger.child({
    redisUrl,
    options,
    source: 'redis-extension',
  })
  logr.debug({ redisUrl, options }, 'connecting to redis...')
  if (!options.shouldReconnect) {
    options.retry_strategy = () => undefined
  }
  const redisClient = redis.createClient(redisUrl, options)

  const hasConnected = new Promise((resolve, reject) => {
    if (redisClient.ready) {
      resolve(redisClient)
      return
    }

    redisClient.on('ready', () => {
      logr.debug('Connection to redis has been established successfully.')
      resolve(redisClient)
    })

    redisClient.on('error', (err) => {
      logr.error({ err }, 'Redis error connecting.')
      reject(err)
    })

    redisClient.on('end', () => {
      logr.error('Redis connection closed.')
      reject(new Error('Redis connection closed.'))
    })
  })

  await hasConnected

  const result = await redisClient.pingAsync()
  if (!result) {
    throw new Error('Failed to get server status from redis.')
  }
  logr.info({ redisUrl }, 'Successfully connected to redis.')
  return redisClient
}

export async function withCriticalSection(redisClient, f) {
  if (!Lock) {
    const options = { retryCount: 50, retryDelay: 10 }
    Lock = new Redlock([redisClient], options)
  }
  const rlock = await Lock.lock(LockKey, LockTTL)
  const res = await f()
  await rlock.unlock()
  return res
}

export async function disconnect(client) {
  const hasDisconnected = new Promise((resolve, reject) => {
    client.quit((err, res) => {
      if (err) {
        reject(err)
        return
      }
      resolve(res)
    })
  })

  await hasDisconnected
}
