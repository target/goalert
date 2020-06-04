export function pageFab(dialOption?: string): Cypress.Chainable {
  // standard page fab
  if (!dialOption)
    return cy.get('button[data-cy=page-fab]').should('be.visible').click()

  // speed dial page fab
  return cy
    .get('button[data-cy=page-fab]')
    .should('be.visible')
    .trigger('mouseover')
    .parent()
    .find(
      `span[aria-label*=${JSON.stringify(dialOption)}] button[role=menuitem]`,
    )
    .click()
}

Cypress.Commands.add('pageFab', pageFab)
