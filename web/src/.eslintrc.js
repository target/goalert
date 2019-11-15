const fs = require('fs')
const path = require('path')

module.exports = {
  parser: 'babel-eslint',
  parserOptions: {
    ecmaFeatures: { legacyDecorators: true },
  },
  plugins: ['cypress', 'prettier', 'jsx-a11y', 'react-hooks'],
  extends: [
    'standard',
    'standard-jsx',
    'plugin:cypress/recommended',
    'plugin:prettier/recommended',
    'plugin:jsx-a11y/recommended',
  ],
  rules: {
    'prettier/prettier': 'error',
    'react-hooks/rules-of-hooks': 'error',
  },
  env: {
    'cypress/globals': true,
  },
  globals: {
    beforeAll: 'readonly',
    afterAll: 'readonly',
  },
}
