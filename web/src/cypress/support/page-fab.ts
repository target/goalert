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
  const res = cy.get('button[data-cy=page-fab]').click()
  if (!dialOption) return res

  return res
    .parent()
    .find(`button[role=menuitem][aria-label~=${JSON.stringify(dialOption)}]`)
    .click()
}

Cypress.Commands.add('pageFab', pageFab)
