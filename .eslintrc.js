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

    '@typescript-eslint/no-empty-function': 'off',
    '@typescript-eslint/explicit-function-return-type': [
      'error',
      {
        allowExpressions: true,
        allowTypedFunctionExpressions: false,
      },
    ],
    '@typescript-eslint/no-unused-vars': [
      'error',
      {
        ignoreRestSiblings: true,
      },
    ],
    'no-use-before-define': 'off',
    '@typescript-eslint/no-use-before-define': ['error'],
    'array-callback-return': 'off',
  },
  overrides: [
    {
      files: ['*.js', '*.jsx'],
      rules: {
        '@typescript-eslint/explicit-function-return-type': 'off',
        '@typescript-eslint/explicit-module-boundary-types': 'off',
      },
    },
    {
      files: ['*.d.ts'],
      rules: {
        '@typescript-eslint/no-unused-vars': 'off',
      },
    },
  ],
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
