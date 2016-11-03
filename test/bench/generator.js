// The following is not needed to create a session file. We don't want to
// re-create & re-allocate memory every time we receive a message so we cache
// them in a variable.
var cache = Object.create(null);

exports.utf8 = function utf(size, fn) {
  if (!cache.joinBody) {
    cache.joinBody = JSON.stringify({
      route: 'matchmaking.join',
      args: 'hello'
    })
    cache.buffer = new Buffer(cache.joinBody)
  }

  return fn(undefined, cache.buffer)
}
