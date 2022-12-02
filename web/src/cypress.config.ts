import { defineConfig } from "cypress";

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

  blockHosts: ["gravatar.com"],

  e2e: {
    // We've imported your old cypress plugins here.
    // You may want to clean this up later by importing these.
    setupNodeEvents(on, config) {
      return require("./cypress/plugins/index.js")(on, config);
    },
    baseUrl: "http://localhost:3030",
    excludeSpecPattern: "*.map",
  },

  component: {
    devServer: {
      framework: "react",
      bundler: "webpack",
    },
  },
});
