interface MenuSelectOptions {
  /** Forces the menu to operate in widescreen mode.
   *
   * Useful on pages that haven't been made mobile friendly yet.
   */
  forceWidescreen?: boolean
}

declare global {
  namespace Cypress {
    interface Chainable {
      /** Open the selected menu and click the matching item. */
      menu: (label: string, options?: MenuSelectOptions) => Cypress.Chainable
    }
  }
}

function menu(sub: JQuery<HTMLElement>, s: string): Cypress.Chainable {
  return cy.get('[data-cy=app-bar]').then((el) => {
    const format: 'mobile' | 'wide' = el.data('cy-format')
    expect(format, 'header format').to.be.oneOf(['mobile', 'wide'])

    // open menu
    cy.wrap(sub).click()

    // click menu item
    cy.get('ul[role=menu]').contains('[role=menuitem]', s).click()
    cy.get('ul[role=menu]').should('not.be.visible')
  })
}

Cypress.Commands.add('menu', { prevSubject: 'element' }, menu)

export {}
