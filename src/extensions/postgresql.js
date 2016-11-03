import Sequelize from 'sequelize'

export async function check(db) {
  const result = {
    up: false,
  }

  try {
    const query = "select * from pg_stat_activity where state='active';"
    const deadlockQuery = 'SELECT blocked_locks.pid     AS blocked_pid, ' +
           'blocked_locks.virtualtransaction as blocked_transaction_id, ' +
           'blocked_activity.usename  AS blocked_user, ' +
           'blocking_locks.pid     AS blocking_pid, ' +
           'blocking_locks.virtualtransaction as blocking_transaction_id, ' +
           'blocking_activity.usename AS blocking_user, ' +
           'blocked_activity.query    AS blocked_statement, ' +
           'blocking_activity.query   AS blocking_statement ' +
        'FROM  pg_catalog.pg_locks         blocked_locks ' +
        'JOIN pg_catalog.pg_stat_activity blocked_activity  ON blocked_activity.pid = blocked_locks.pid ' +
        'JOIN pg_catalog.pg_locks         blocking_locks  ' +
            'ON blocking_locks.locktype = blocked_locks.locktype ' +
            'AND blocking_locks.DATABASE IS NOT DISTINCT FROM blocked_locks.DATABASE ' +
            'AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation ' +
            'AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page ' +
            'AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple ' +
            'AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid ' +
            'AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid ' +
            'AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid ' +
            'AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid ' +
            'AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid ' +
            'AND blocking_locks.pid != blocked_locks.pid ' +
        'JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid ' +
        'WHERE NOT blocked_locks.GRANTED; '
    const res = await db.query(query, { type: Sequelize.QueryTypes.SELECT })
    if (res) {
      result.up = true
      result.activeOperations = res.length
      const deadlock = await db.query(deadlockQuery, { type: Sequelize.QueryTypes.SELECT })
      result.deadlock = deadlock.length === 0
      result.deadlockOperations = []
      deadlock.forEach((row) => {
        result.deadlockOperations.push({
          blocked: {
            pid: row.blocked_pid,
            txId: row.blocked_transaction_id,
            user: row.blocked_user,
            statement: row.blocked_statement,
          },
          blocking: {
            pid: row.blocking_pid,
            txId: row.blocking_transaction_id,
            user: row.blocking_user,
            statement: row.blocking_statement,
          },
        })
      })
    } else {
      result.error = 'Could not get server status!'
    }
  } catch (error) {
    result.error = error.message
  }

  return result
}

export async function connect(pgUrl, options, logger) {
  let opt = options
  if (!options) {
    opt = {}
  }

  opt.dialect = 'postgres'

  logger.debug({ pgUrl, opt }, 'Connecting to PostgreSQL...')

  try {
    const db = new Sequelize(pgUrl, opt)
    const query = 'select 1;'
    const res = await db.query(query, { type: Sequelize.QueryTypes.SELECT })
    if (!res) {
      const err = new Error('Failed to connect to PostgreSQL.')
      logger.error({ pgUrl, err }, err.message)
      throw err
    }

    logger.info({ pgUrl }, 'Successfully connected to PostgreSQL.')
    return db
  } catch (err) {
    logger.error({ pgUrl, err }, 'Failed to connect to PostgreSQL.')
    throw err
  }
}
