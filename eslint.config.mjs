import { fixupConfigRules, fixupPluginRules } from '@eslint/compat'
import _import from 'eslint-plugin-import'
import cypress from 'eslint-plugin-cypress'
import jsxA11Y from 'eslint-plugin-jsx-a11y'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import prettier from 'eslint-plugin-prettier'
import typescriptEslint from '@typescript-eslint/eslint-plugin'
import tsParser from '@typescript-eslint/parser'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import js from '@eslint/js'
import { FlatCompat } from '@eslint/eslintrc'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const compat = new FlatCompat({
  baseDirectory: __dirname,
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
})

export default [
  {
    ignores: [
      'web/src/build',
      'web/src/profile.json',
      'bin',
      'storybook-static',
      'playwright/.cache',
      'web/src/app/editor/expr-parser.ts',
    ],
  },
  ...fixupConfigRules(
    compat.extends(
      'standard',
      'standard-jsx',
      'plugin:import/errors',
      'plugin:import/warnings',
      'plugin:import/typescript',
      'plugin:react/recommended',
      'plugin:cypress/recommended',
      'plugin:jsx-a11y/recommended',
      'plugin:prettier/recommended',
      'plugin:@typescript-eslint/eslint-recommended',
      'plugin:@typescript-eslint/recommended',
      'prettier',
      'plugin:storybook/recommended',
    ),
  ),
  {
    plugins: {
      import: fixupPluginRules(_import),
      cypress: fixupPluginRules(cypress),
      'jsx-a11y': fixupPluginRules(jsxA11Y),
      react: fixupPluginRules(react),
      'react-hooks': fixupPluginRules(reactHooks),
      prettier: fixupPluginRules(prettier),
      '@typescript-eslint': fixupPluginRules(typescriptEslint),
    },

    languageOptions: {
      globals: {
        ...cypress.environments.globals.globals,
        beforeAll: 'readonly',
        afterAll: 'readonly',
      },

      parser: tsParser,
    },

    settings: {
      react: {
        version: 'detect',
      },
    },

    rules: {
      'no-else-return': [
        'error',
        {
          allowElseIf: false,
        },
      ],

      'prettier/prettier': 'error',
      'react-hooks/rules-of-hooks': 'error',
      'react/jsx-fragments': ['error', 'element'],
      'react/prop-types': 'off',
      'no-useless-constructor': 'off',
      'import/no-unresolved': 'off',
      'react/jsx-curly-newline': 'off',
      'react/jsx-indent': 'off',
      '@typescript-eslint/no-namespace': 'off',
      '@typescript-eslint/no-empty-function': 'off',

      '@typescript-eslint/explicit-function-return-type': [
        'error',
        {
          allowExpressions: true,
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
  },
  {
    files: ['**/*.js', '**/*.jsx'],

    rules: {
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off',
    },
  },
  {
    files: ['**/*.d.ts'],

    rules: {
      '@typescript-eslint/no-unused-vars': 'off',
    },
  },
]
