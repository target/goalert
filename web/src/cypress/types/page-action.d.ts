declare namespace Cypress {
  interface Chainable {
    /** Click an action from the page-level "Other Actions" menu. */
    pageAction: typeof pageAction
  }
}
