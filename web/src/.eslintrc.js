const fs = require('fs')
const path = require('path')

module.exports = {
  parser: 'babel-eslint',
  parserOptions: {
    ecmaFeatures: { legacyDecorators: true },
  },
  plugins: ['cypress', 'prettier'],
  extends: [
    'standard',
    'standard-jsx',
    'plugin:cypress/recommended',
    'plugin:prettier/recommended',
  ],
  rules: {
    'prettier/prettier': 'error',
  },
  env: {
    'cypress/globals': true,
  },
  globals: {
    beforeAll: 'readonly',
    afterAll: 'readonly',
  },
}
