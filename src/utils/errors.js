export class InvalidRouteError extends Error {

  constructor(message, extra) {
    super(message)
    this.name = this.constructor.name
    this.message = message
    if (typeof Error.captureStackTrace === 'function') {
      Error.captureStackTrace(this, this.constructor)
    } else {
      this.stack = (new Error(message)).stack
    }
    this.extra = extra
    this.code = 11000
  }

}

export class RouteNotFoundError extends Error {

  constructor(message, extra) {
    super(message)
    this.name = this.constructor.name
    this.message = message
    if (typeof Error.captureStackTrace === 'function') {
      Error.captureStackTrace(this, this.constructor)
    } else {
      this.stack = (new Error(message)).stack
    }
    this.extra = extra
    this.code = 11001
  }

}

export class InvalidArgumentsError extends Error {
  constructor(message, extra) {
    super(message)
    this.name = this.constructor.name
    this.message = message
    if (typeof Error.captureStackTrace === 'function') {
      Error.captureStackTrace(this, this.constructor)
    } else {
      this.stack = (new Error(message)).stack
    }
    this.extra = extra
    this.code = 11002
  }
}

export class PlayerIdNotDefinedError extends Error {
  constructor(message, extra) {
    super(message)
    this.name = this.constructor.name
    this.message = message
    if (typeof Error.captureStackTrace === 'function') {
      Error.captureStackTrace(this, this.constructor)
    } else {
      this.stack = (new Error(message)).stack
    }
    this.extra = extra
    this.code = 11003
  }
}

export class GenericError extends Error {
  constructor(message, extra) {
    super(message)
    this.name = this.constructor.name
    this.message = message
    if (typeof Error.captureStackTrace === 'function') {
      Error.captureStackTrace(this, this.constructor)
    } else {
      this.stack = (new Error(message)).stack
    }
    this.extra = extra
    this.code = 11003
  }
}
