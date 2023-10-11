declare global {
  namespace Cypress {
    interface Chainable {
      /** Enter a page-level search (from the top bar). Works in mobile and widescreen. */
      pageSearch: typeof pageSearch
    }
  }
}

function pageSearch(s: string): Cypress.Chainable {
  return cy.get('[data-cy=app-bar]').then((el) => {
    const format: 'mobile' | 'wide' = el.data('cy-format')
    expect(format, 'header format').to.be.oneOf(['mobile', 'wide'])

    if (format === 'mobile') {
      cy.get('[data-cy=app-bar] button[data-cy=open-search]').click({
        // since we're running tests, it's ok if it is already open
        force: true,
      })
      cy.get('input[name=search]').type(`{selectall}${s}`, {
        // work around bug with search/app-bar where it doesn't register the keypress
        // TODO: move ownership to app bar instead of container magic
        delay: 10,
      })
    } else {
      cy.form({ search: s })
    }

    cy.get('input[name=search]').should('have.value', s)
  })
}

Cypress.Commands.add('pageSearch', pageSearch)

export {}
