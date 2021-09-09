/* eslint @typescript-eslint/no-var-requires: 0 */
const path = require('path')
const webpack = require('webpack')
const CopyPlugin = require('copy-webpack-plugin')

// Constants
const APP = path.join(__dirname, 'app')
const BUILD = path.join(__dirname, 'build')
const CYPRESS = path.join(__dirname, 'cypress')

module.exports = (env) => ({
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
        use: ['babel-loader'],
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
  devtool: 'eval-source-map',
  // webpack-dev-server configuration
  devServer: {
    port: 3035,
    allowedHosts: ['docker.for.mac.host.internal'],
    watchFiles: [APP, CYPRESS],

    devMiddleware: {
      stats: 'errors-only',
    },
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
