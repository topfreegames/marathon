import { expect } from './common'

describe('Worker', () => {
  it.only('should create worker', async function () {
    expect(this.worker).not.to.be.null()
  })
})
