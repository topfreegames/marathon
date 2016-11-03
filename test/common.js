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
