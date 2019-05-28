const modules = process.env.NODE_ENV === 'test' ? 'cjs' : false
module.exports = {
  presets: ['@babel/preset-react', ['@babel/preset-env', { modules }]],
  plugins: [
    '@babel/plugin-transform-runtime',
    '@babel/plugin-syntax-dynamic-import',
    ['@babel/plugin-proposal-decorators', { legacy: true }],
    ['@babel/plugin-proposal-class-properties', { loose: true }],
  ],
}
