declare global {
  namespace Cypress {
    interface Chainable {
      /** Click an action from the page-level "Other Actions" menu. */
      pageAction: typeof pageAction
    }
  }
}

function pageAction(s: string): Cypress.Chainable {
  return cy
    .get('[data-cy=app-bar]')
    .find('button[data-cy=other-actions]')
    .menu(s)
}

Cypress.Commands.add('pageAction', pageAction)

export {}
