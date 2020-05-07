export function pageAction(s: string): Cypress.Chainable {
  return cy
    .get('[data-cy=app-bar]')
    .find('button[data-cy=other-actions]')
    .menu(s)
}

Cypress.Commands.add('pageAction', pageAction)
