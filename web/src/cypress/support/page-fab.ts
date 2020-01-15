declare namespace Cypress {
  interface Chainable {
    /** Click the FAB (floating action button) of the page.
     *
     * If the FAB is a Speed-Dial variant, you can optionally
     * specify the option label to select as an argument.
     */
    pageFab: typeof pageFab
  }
}

function pageFab(dialOption?: string): Cypress.Chainable {
  const res = cy.get('button[data-cy=page-fab]').should('be.visible')

  // standard page fab
  if (!dialOption) return res.click()

  // speed dial page fab
  return res
    .trigger('mouseover')
    .parent()
    .find(
      `span[aria-label*=${JSON.stringify(dialOption)}] button[role=menuitem]`,
    )
    .click()
}

Cypress.Commands.add('pageFab', pageFab)
