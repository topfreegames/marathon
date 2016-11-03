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
  logger.debug({ redisUrl, options }, 'connecting to redis...')
  if (!options.shouldReconnect) {
    options.retry_strategy = () => undefined
  }
  const redisClient = redis.createClient(redisUrl, options)

  redisClient.on('error', (err) => {
    logger.error({ err }, 'redis error')
  })

  const result = await redisClient.pingAsync()
  if (!result) {
    throw new Error('Failed to get server status from redis.')
  }
  logger.info({ redisUrl }, 'successfully connected to redis')
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
