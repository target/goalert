/* eslint @typescript-eslint/no-var-requires: 0 */
// Folder ops
const path = require('path')
const webpack = require('webpack')
const CopyPlugin = require('copy-webpack-plugin')

// Constants
const APP = path.join(__dirname, 'app')
const BUILD = path.join(__dirname, 'build')
const PORT = process.env.PORT || 3032

module.exports = (env = { GOALERT_VERSION: 'dev' }) => ({
  mode: 'development',
  // Paths and extensions
  entry: {
    app: APP,
  },
  cache: {
    type: 'filesystem',
  },
  output: {
    path: BUILD,
    filename: 'static/[name].js',
  },
  resolve: {
    extensions: ['.mjs', '.ts', '.tsx', '.js', '.jsx', '.css'],
  },
  // Loaders for processing different file types
  module: {
    rules: [
      { test: /lodash/, loader: 'strict-loader' },
      {
        test: /\.(t|j)sx?$/,
        use: [
          'babel-loader',
          { loader: 'ifdef-loader', options: { production: false } },
        ],
        include: [APP],
      },
      {
        test: /\.css$/,
        use: [
          'style-loader',
          { loader: 'css-loader', options: { importLoaders: 1 } },
        ],
      },
      {
        test: /\.json$/,
        loader: 'json-loader',
        type: 'javascript/auto',
      },
      {
        test: /\.md$/,
        type: 'asset/source',
      },
      {
        test: /\.(gif|png|jpe?g|svg|ico|webp)$/i,
        type: 'asset/resource',
        generator: {
          filename: 'static/[hash].[ext]',
        },
      },
    ],
  },

  // Source maps used for debugging information
  devtool: 'inline-source-map',
  // webpack-dev-server configuration
  devServer: {
    disableHostCheck: true,

    stats: 'errors-only',

    // host: HOST,
    port: PORT,
  },
  optimization: {
    splitChunks: {
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendor',
          chunks: 'all',
        },
      },
    },
  },

  // Webpack plugins
  plugins: [
    new webpack.DefinePlugin({
      'process.env': {
        NODE_ENV: JSON.stringify('development'), // eslint-disable-line quote-props
      },
    }),
    new CopyPlugin({
      patterns: [
        'favicon-16.png',
        'favicon-32.png',
        'favicon-64.png',
        'favicon-192.png',
        'goalert-alt-logo.png',
      ].map((filename) => ({
        from: path.resolve(APP, `./public/${filename}`),
        to: path.resolve(BUILD, `./static/${filename}`),
      })),
    }),
    new webpack.BannerPlugin({
      banner: `var GOALERT_VERSION=${JSON.stringify(env.GOALERT_VERSION)};`,
      raw: true,
    }),
  ],
})
