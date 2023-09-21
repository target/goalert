/* eslint-disable @typescript-eslint/no-var-requires */

const { defineConfig } = require('cypress')
const setupNodeEvents = require('./web/src/cypress/plugins/index')

module.exports = defineConfig({
  videoUploadOnPasses: false,
  waitForAnimations: false,
  viewportWidth: 1440,
  viewportHeight: 900,
  requestTimeout: 15000,
  defaultCommandTimeout: 15000,
  video: false,

  retries: {
    runMode: 2,
    openMode: 0,
  },

  blockHosts: ['gravatar.com'],

  e2e: {
    setupNodeEvents,
    baseUrl: 'http://localhost:3030',
    excludeSpecPattern: '*.map',
    supportFile: 'bin/build/integration/cypress/support/index.js',
    specPattern: 'bin/build/integration/cypress/integration/*.cy.js',
  },

  component: {
    devServer: {
      framework: 'react',
      bundler: 'webpack',
    },
  },
})
