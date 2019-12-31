declare namespace Cypress {
  interface Chainable<Subject> {
    /** Click a dialog button with the given text and wait for it to dissapear. */
    dialogFinish: typeof dialogFinish
  }
}

function dialogFinish(s: string): Cypress.Chainable {
  return cy
    .get('[role=dialog]')
    .should('be.visible')
    .contains('button', s)
    .click()
    .get('[role=dialog]')
    .should('not.exist')
}

Cypress.Commands.add('dialogFinish', dialogFinish)
