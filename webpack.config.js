// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

const path = require('path')
const fs = require('fs')

const nodeModules = {}
fs.readdirSync('node_modules')
  .filter(x =>
    ['.bin'].indexOf(x) === -1
  )
  .forEach((mod) => {
    nodeModules[mod] = `commonjs ${mod}`
  })

module.exports = {
  entry: {
    marathon: ['babel-polyfill', './src/index.js'],
  },
  target: 'node',
  module: {
    loaders: [{
      test: /\.jsx?$/,
      exclude: /node_modules/,
      loader: 'babel',
    }, {
      test: /\.json$/,
      exclude: /node_modules/,
      loader: 'json',
    }],
  },
  output: {
    path: path.join(__dirname, 'lib'),
    filename: '[name].js',
  },
  externals: nodeModules,
  devtool: 'sourcemap',
}
