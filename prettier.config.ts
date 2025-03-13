 
const pluginGoTemplate = require.resolve('prettier-plugin-go-template')
module.exports = {
  trailingComma: 'all',
  semi: false,
  jsxSingleQuote: true,
  singleQuote: true,
  endOfLine: 'lf',
  plugins: [pluginGoTemplate],
  overrides: [
    {
      files: ['*.html'],
      options: {
        parser: 'go-template',
      },
    },
  ],
}
