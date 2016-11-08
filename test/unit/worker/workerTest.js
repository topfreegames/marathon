// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import { expect } from './common'

describe('Worker', () => {
  it('should create worker', async function () {
    expect(this.worker).not.to.be.null()
  })
})
