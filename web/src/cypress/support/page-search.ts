declare namespace Cypress {
  interface Chainable {
    /** Enter a page-level search (from the top bar). Works in mobile and widescreen. */
    pageSearch: typeof pageSearch
  }
}

function pageSearch(s: string): Cypress.Chainable {
  cy.get('[data-cy=app-bar]').as('container')

  return cy.get('@container').then(el => {
    const format: 'mobile' | 'wide' = el.data('cy-format')
    expect(format, 'header format').to.be.oneOf(['mobile', 'wide'])

    if (format === 'mobile') {
      cy.get('@container')
        .find('button[data-cy=open-search]')
        .click()
    }

    cy.get('@container')
      .find('input')
      .type(`{selectall}${s}{enter}`)

    if (format === 'mobile') {
      cy.get('@container')
        .find('button[data-cy=close-search]')
        .click()
    }
  })
}

Cypress.Commands.add('pageSearch', pageSearch)
