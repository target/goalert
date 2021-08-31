declare namespace Cypress {
  interface Chainable {
    /** Gets a list of slack channels */
    getSlackChannels: () => Cypress.Chainable<SlackChannel[]>
  }
}

interface SlackChannel {
  id: string
  name: string
}
