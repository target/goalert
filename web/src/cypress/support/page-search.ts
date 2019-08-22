declare namespace Cypress {
  interface Chainable {
    /** Enter a page-level search (from the top bar). Works in mobile and widescreen. */
    pageSearch: typeof pageSearch
  }
}

function pageSearch(s: string): Cypress.Chainable {
  return cy.get('[data-cy=app-bar]').then(el => {
    const format: 'mobile' | 'wide' = el.data('cy-format')
    expect(format, 'header format').to.be.oneOf(['mobile', 'wide'])

    if (format === 'mobile') {
      cy.get('[data-cy=app-bar] button[data-cy=open-search]').click({
        force: true,
      }) // since we're running tests, it's ok if it is already open
    }

    cy.get('[data-cy=app-bar] input')
      .type(`{selectall}${s}`)
      .should('have.value', s)
  })
}

Cypress.Commands.add('pageSearch', pageSearch)
