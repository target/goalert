const path = require('path')
const glob = require('glob')

module.exports = {
  mode: 'production',
  // Paths and extensions
  entry: {
    'support/index': path.join(__dirname, 'cypress/support'),
    'integration/all': glob.sync(path.join(__dirname, 'cypress/integration/*')),
  },
  output: {
    path: path.join(__dirname, '../../bin/integration/goalert/cypress'),
  },
  resolve: {
    extensions: ['.js', '.ts'],
  },
  // Loaders for processing different file types
  module: {
    rules: [
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
            loader: 'ts-loader',
          },
        ],
      },
    ],
  },

  // Source maps used for debugging information
  devtool: 'cheap-module-eval-source-map',
}
