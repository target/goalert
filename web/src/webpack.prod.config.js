// Initialization
const webpack = require('webpack')

// File ops
const HtmlWebpackPlugin = require('html-webpack-plugin')

// Folder ops
const path = require('path')

// Constants
const APP = path.join(__dirname, 'app')
const BUILD = path.join(__dirname, 'build')
const TEMPLATE = path.join(__dirname, 'app/templates/index.html')

module.exports = {
  mode: 'production',
  entry: {
    app: APP,
  },
  output: {
    path: BUILD,
    filename: 'static/[name].[chunkhash].js',
    chunkFilename: 'static/[chunkhash].js',
    publicPath: '/',
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
      },
    }),
    // Auto generate index.html
    new HtmlWebpackPlugin({
      // custom favicon
      favicon: 'app/public/favicon.ico',
      template: TEMPLATE,
      // JS placed at the bottom of the body element
      inject: 'body',
      // Use html-minifier
      minify: {
        collapseWhitespace: true,
      },
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
}
