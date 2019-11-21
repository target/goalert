const fs = require('fs')
const path = require('path')

module.exports = {
  parser: 'babel-eslint',
  parserOptions: {
    ecmaFeatures: { legacyDecorators: true },
  },
  plugins: ['cypress', 'jsx-a11y', 'react-hooks', 'prettier'],
  extends: [
    'standard',
    'standard-jsx',
    'plugin:cypress/recommended',
    'plugin:jsx-a11y/recommended',
    'plugin:prettier/recommended',
  ],
  rules: {
    'no-else-return': ['error', { allowElseIf: false }],
    'prettier/prettier': 'error',
    'react-hooks/rules-of-hooks': 'error',
    'react/jsx-fragments': ['error', 'element'],

    // handled by prettier
    'react/jsx-curly-newline': 'off',
    'react/jsx-indent': 'off',
  },
  settings: {
    react: { version: 'detect' },
  },
  env: {
    'cypress/globals': true,
  },
  globals: {
    beforeAll: 'readonly',
    afterAll: 'readonly',
  },
}
