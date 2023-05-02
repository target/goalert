/* eslint @typescript-eslint/no-var-requires: 0 */
const path = require('path')
const glob = require('glob')

module.exports = {
  mode: 'production',
  // Paths and extensions
  entry: {
    'support/e2e': path.join(__dirname, 'cypress/support'),
    'integration/all': glob.globSync(
      path.join(__dirname, 'cypress/integration/*'),
    ),
    'plugins/index': path.join(__dirname, 'cypress/ci-plugins.js'),
  },
  target: 'node',
  output: {
    path: path.join(__dirname, '../../bin/build/integration/cypress'),
    libraryTarget: 'commonjs-module',
    libraryExport: 'default',
  },
  resolve: {
    extensions: ['.js', '.ts'],
  },
  // Loaders for processing different file types
  module: {
    rules: [
      { test: /lodash/, loader: 'strict-loader' },
      {
        test: /\.json$/,
        loader: 'json-loader',
        type: 'javascript/auto',
      },
      {
        test: /\.ts$/,
        exclude: [/node_modules/],
        use: [
          {
            loader: 'babel-loader',
          },
        ],
      },
    ],
  },

  // omit source maps for maximum performance; turn on for debugging
  // devtool: 'eval-cheap-module-source-map',
}
