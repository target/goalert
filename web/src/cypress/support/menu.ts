declare namespace Cypress {
  interface Chainable<Subject> {
    /** Open the selected menu and click the matching item. */
    menu: menuFn
  }
}

interface MenuSelectOptions {
  /** Forces the menu to operate in widescreen mode.
   *
   * Useful on pages that haven't been made mobile friendly yet.
   */
  forceWidescreen?: Boolean
}

type menuFn = (label: string, options?: MenuSelectOptions) => Cypress.Chainable

function menu(
  sub: any,
  s: string,
  options?: MenuSelectOptions,
): Cypress.Chainable {
  return cy.get('[data-cy=app-bar]').then(el => {
    const format: 'mobile' | 'wide' = el.data('cy-format')
    expect(format, 'header format').to.be.oneOf(['mobile', 'wide'])

    // Can't just use `.click()` due to
    // an issue with Cypress 3.6.0
    cy.wrap(sub)
      .eq(0)
      .then(el => el.click())

    if ((options && options.forceWidescreen) || format === 'wide') {
      cy.get('ul[role=menu]')
        .contains('li', s)
        .click()
    } else {
      cy.get('ul[data-cy=mobile-actions]')
        .contains('*[role=button]', s)
        .click()
    }
  })
}

Cypress.Commands.add('menu', { prevSubject: 'element' }, menu)
