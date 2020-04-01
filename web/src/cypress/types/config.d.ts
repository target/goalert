declare namespace Cypress {
  interface Chainable {
    getConfig: typeof getConfig

    /** Replaces the backend config entirely. */
    setConfig: typeof setConfig

    /** Merges new config values into existing backend config. */
    updateConfig: typeof updateConfig

    resetConfig: typeof resetConfig
  }
}
