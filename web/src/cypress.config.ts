import { defineConfig } from 'cypress'
import setupNodeEvents from './cypress/plugins/index'

export default defineConfig({
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
  },

  component: {
    devServer: {
      framework: 'react',
      bundler: 'webpack',
    },
  },
})
