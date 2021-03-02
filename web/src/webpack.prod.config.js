/* eslint @typescript-eslint/no-var-requires: 0 */
// Initialization
const webpack = require('webpack')
const CopyPlugin = require('copy-webpack-plugin')

// Folder ops
const path = require('path')

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

  // Remove comment if you require sourcemaps for your production code
  devtool: 'source-map',
  plugins: [
    // Required to inject NODE_ENV within React app.
    // Redundant package.json script entry does not do that, but required for .babelrc
    // Optimizes React for use in production mode
    new webpack.DefinePlugin({
      'process.env': {
        NODE_ENV: JSON.stringify('production'), // eslint-disable-line quote-props
        GOALERT_VERSION: JSON.stringify(env.GOALERT_VERSION), // eslint-disable-line quote-props
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
