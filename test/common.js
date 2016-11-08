// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import chaiMod from 'chai'
import dirtyChai from 'dirty-chai'

const initializeChai = () => {
  chaiMod.config.includeStack = true // turn on stack trace
  chaiMod.use(dirtyChai)
}

const initializeTests = () => {
  process.env.NODE_ENV = 'test'

  initializeChai()
}

export const expect = chaiMod.expect
export const chai = chaiMod

initializeTests()
