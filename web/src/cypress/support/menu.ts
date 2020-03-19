interface MenuSelectOptions {
  /** Forces the menu to operate in widescreen mode.
   *
   * Useful on pages that haven't been made mobile friendly yet.
   */
  forceWidescreen?: boolean
}

function menu(
  sub: any,
  s: string,
  options?: MenuSelectOptions,
): Cypress.Chainable {
  return cy.get('[data-cy=app-bar]').then(el => {
    const format: 'mobile' | 'wide' = el.data('cy-format')
    expect(format, 'header format').to.be.oneOf(['mobile', 'wide'])

    // open menu
    cy.wrap(sub).click()

    // click menu item
    if ((options && options.forceWidescreen) || format === 'wide') {
      cy.get('ul[role=menu]')
        .contains('li', s)
        .click()
      cy.get('ul[role=menu]').should('not.exist')
    } else {
      cy.get('ul[data-cy=mobile-actions]')
        .contains('*[role=button]', s)
        .click()
      cy.get('ul[data-cy=mobile-actions]').should('not.exist')
    }
  })
}

Cypress.Commands.add('menu', { prevSubject: 'element' }, menu)

export {}
