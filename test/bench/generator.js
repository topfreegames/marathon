// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

// The following is not needed to create a session file. We don't want to
// re-create & re-allocate memory every time we receive a message so we cache
// them in a variable.
const cache = Object.create(null)

exports.utf8 = function utf(size, fn) {
  if (!cache.joinBody) {
    cache.joinBody = JSON.stringify({
      route: 'matchmaking.join',
      args: 'hello',
    })
    cache.buffer = new Buffer(cache.joinBody)
  }

  return fn(undefined, cache.buffer)
}
