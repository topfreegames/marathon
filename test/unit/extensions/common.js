// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import { expect as exp, chai as chaiMod } from '../../common'
import config from 'config'  //eslint-disable-line
import Logger from '../../../src/extensions/logger'

export const expect = exp
export const chai = chaiMod

export async function beforeEachFunc(self) {
  self.config = config
  self.logger = new Logger(self.config).logger
}
