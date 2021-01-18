/* eslint @typescript-eslint/no-var-requires: 0 */
const webpack = require('webpack')
const path = require('path')
const outputPath = path.join(__dirname, 'build')
const TerserPlugin = require('terser-webpack-plugin')

module.exports = {
  entry: {
    // All infrequently changed packages
    vendorPackages: [
      'react',
      'react-dom',
      'react-router',
      'react-router-dom',
      '@material-ui/core',
      '@material-ui/icons',
      '@material-ui/lab',
      '@material-ui/pickers',
      'luxon',
      'reselect',
      '@apollo/client',
      'redux',
      'redux-devtools-extension',
      'redux-thunk',
      'react-redux',
      'mdi-material-ui',
      '@date-io/luxon',
      'classnames',
      'react-beautiful-dnd',
      'react-ga',
      'history',
      'react-countdown-now',
      'connected-react-router',
      'chance',
      'ansi-html',
      'ansi-regex',
      'date-arithmetic',
      'diff',
      'dom-helpers',
      'events',
      'fast-levenshtein',
      'html-entities',
      'lodash',
      'loglevel',
      'react-big-calendar',
      'react-infinite-scroll-component',
      'react-overlays',
      'sockjs-client',
      'strip-ansi',
      'uncontrollable',
      'url',
    ],
  },
  mode: 'development',
  output: {
    filename: '[name].dll.js',
    path: outputPath,
    library: '[name]',
  },

  plugins: [
    new webpack.DllPlugin({
      name: '[name]',
      path: path.join(outputPath, '[name].json'),
    }),
  ],
  optimization: {
    // minify javascript
    minimizer: [new TerserPlugin()],
  },
}
