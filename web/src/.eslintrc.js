const fs = require('fs')
const path = require('path')

module.exports = {
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaFeatures: { legacyDecorators: true },
  },
  plugins: [
    'cypress',
    'jsx-a11y',
    'react-hooks',
    'prettier',
    '@typescript-eslint',
  ],
  extends: [
    'standard',
    'standard-jsx',
    'plugin:cypress/recommended',
    'plugin:jsx-a11y/recommended',
    'plugin:prettier/recommended',
    'plugin:@typescript-eslint/eslint-recommended',
    'plugin:@typescript-eslint/recommended',
    'prettier/@typescript-eslint',
  ],
  rules: {
    'no-else-return': ['error', { allowElseIf: false }],
    'prettier/prettier': 'error',
    'react-hooks/rules-of-hooks': 'error',
    'react/jsx-fragments': ['error', 'element'],

    // handled by prettier
    'react/jsx-curly-newline': 'off',
    'react/jsx-indent': 'off',

    // typescript-eslint rules
    // TODO: add options { allowExpressions: true, allowTypedFunctionExpressions: false }
    '@typescript-eslint/explicit-function-return-type': 'off',
    // TODO: use defaults
    '@typescript-eslint/no-namespace': 'off',
    // TODO: use defaults
    '@typescript-eslint/no-use-before-define': 'off',
    // TODO: no-explicit-any: on
    '@typescript-eslint/no-explicit-any': 'off',
    '@typescript-eslint/no-empty-function': 'off',
    // TODO: on, but bypass alerts files
    '@typescript-eslint/camelcase': 'off',
    '@typescript-eslint/no-unused-vars': [
      'error',
      {
        ignoreRestSiblings: true,
      },
    ],
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
