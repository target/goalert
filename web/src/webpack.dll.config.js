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
      'lodash-es',
      'apollo-client',
      'apollo-link',
      'apollo-link-retry',
      'apollo-link-http',
      'apollo-cache-inmemory',
      'reselect',
      'apollo-utilities',
      'axios',
      'redux',
      'redux-devtools-extension',
      'redux-thunk',
      'graphql-tag',
      'react-redux',
      'mdi-material-ui',
      '@date-io/luxon',
      'classnames',
      'react-beautiful-dnd',
      'react-ga',
      'history',
      'react-select',
      'react-apollo',
      'react-countdown-now',
      'connected-react-router',
      'chance',
      '@hot-loader/react-dom',
      'ansi-html',
      'ansi-regex',
      'css-loader',
      'date-arithmetic',
      'diff',
      'dom-helpers',
      'events',
      'fast-levenshtein',
      'html-entities',
      'indexof',
      'lodash',
      'loglevel',
      'querystring-es3',
      'react-big-calendar',
      'react-hot-loader',
      'react-infinite-scroll-component',
      'react-overlays',
      'shallowequal',
      'sockjs-client',
      'strip-ansi',
      'uncontrollable',
      'url',
      '@apollo/react-hooks',
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
