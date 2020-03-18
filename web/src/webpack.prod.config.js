/* eslint @typescript-eslint/no-var-requires: 0 */
// Initialization
const webpack = require('webpack')
const CopyPlugin = require('copy-webpack-plugin')

// Folder ops
const path = require('path')

// Constants
const APP = path.join(__dirname, 'app')
const BUILD = path.join(__dirname, 'build')

module.exports = env => ({
  mode: 'production',
  entry: {
    app: APP,
  },
  output: {
    path: BUILD,
    filename: 'static/[name].js',
  },
  resolve: {
    extensions: ['.ts', '.tsx', '.js', '.jsx', '.css'],
  },
  module: {
    rules: [
      {
        test: /modernizr.config.js$/,
        use: ['modernizr-loader'],
      },
      {
        test: /\.(t|j)sx?$/,
        use: [
          'babel-loader',
          { loader: 'ifdef-loader', options: { production: true, HMR: false } },
        ],
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
        use: 'raw-loader',
      },
      {
        test: /\.(gif|png|jpe?g|svg|ico|webp)$/i,
        use: [
          {
            loader: 'file-loader',
            options: {
              name: 'static/[hash].[ext]',
            },
          },
        ],
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
    new CopyPlugin(
      [16, 32, 64, 192].map(size => ({
        from: path.resolve(APP, `./public/favicon-${size}.png`),
        to: path.resolve(BUILD, `./static/favicon-${size}.png`),
      })),
    ),
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
