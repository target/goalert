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
    'no-else-return': ['error', { allowElseIf: false }],
    'prettier/prettier': 'error',
    'react-hooks/rules-of-hooks': 'error',
    'react/jsx-fragments': ['error', 'element'],
    'react/jsx-curly-newline': [
      'error',
      { multiline: 'consistent', singleline: 'consistent' },
    ],
  },
  env: {
    'cypress/globals': true,
  },
  globals: {
    beforeAll: 'readonly',
    afterAll: 'readonly',
  },
}
