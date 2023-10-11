declare global {
  namespace Cypress {
    interface Chainable {
      /** Click the FAB (floating action button) of the page.
       *
       * If the FAB is a Speed-Dial variant, you can optionally
       * specify the option label to select as an argument.
       */
      pageFab: typeof pageFab
    }
  }
}

function pageFab(dialOption?: string): Cypress.Chainable {
  // standard page fab
  if (!dialOption)
    return cy.get('button[data-cy=page-fab]').should('be.visible').click()

  // speed dial page fab
  cy.get('button[data-cy=page-fab]').should('be.visible').trigger('mouseover')
  return cy
    .get('button[data-cy=page-fab]')
    .parent()
    .find(
      `span[aria-label*=${JSON.stringify(dialOption)}] button[role=menuitem]`,
    )
    .click()
}

Cypress.Commands.add('pageFab', pageFab)

export {}
