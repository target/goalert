declare global {
  namespace Cypress {
    interface Chainable {
      /** Navigate to a link on the side/nav bar. */
      pageNav: typeof pageNav
    }
  }
}

function pageNav(s: string, skipClick?: boolean): Cypress.Chainable {
  return cy.get('*[data-cy=app-bar]').then(el => {
    const format: 'mobile' | 'wide' = el.data('cy-format')
    expect(format, 'app bar format').to.be.oneOf(['mobile', 'wide'])

    if (format === 'mobile') {
      cy.get('button[data-cy=nav-menu-icon]').click({ force: true }) // since we're running tests, it's ok if it's already open
    }
    if (skipClick) {
      cy.get('ul[data-cy=nav-list]').contains('a', s)
    } else {
      cy.get('ul[data-cy=nav-list]')
        .contains('a', s)
        .click()
    }
  })
}

Cypress.Commands.add('pageNav', pageNav)

export {}
