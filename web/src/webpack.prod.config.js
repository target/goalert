/* eslint @typescript-eslint/no-var-requires: 0 */
const path = require('path')
const webpack = require('webpack')
const CopyPlugin = require('copy-webpack-plugin')

// Constants
const APP = path.join(__dirname, 'app')
const BUILD = path.join(__dirname, 'build')

module.exports = (env) => ({
  mode: 'production',
  entry: {
    app: APP,
  },
  output: {
    path: BUILD,
    filename: 'static/[name].js',
  },
  resolve: {
    extensions: ['.mjs', '.ts', '.tsx', '.js', '.jsx', '.css'],
  },
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
          'postcss-loader',
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

  // Comment out for maximum performance; else get high quality SourceMaps
  devtool: 'source-map',

  plugins: [
    new webpack.EnvironmentPlugin({ GOALERT_VERSION: env.GOALERT_VERSION }),
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

  optimization: {
    // separate vendor and manifest files
    splitChunks: {
      chunks(chunk) {
        return chunk.name === 'vendor' || chunk.name === 'manifest'
      },
    },
    // minify javascript
    minimize: true,
  },
})
