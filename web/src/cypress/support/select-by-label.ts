declare namespace Cypress {
  interface Chainable<Subject> {
    /**
     * Selects an item from a dropdown by it's label. Automatically accounts for search-selects.
     */
    selectByLabel: selectByLabelFn

    /**
     * Finds an item from a dropdown by it's label. Automatically accounts for search-selects.
     */
    findByLabel: findByLabelFn

    /**
     * Finds an item from a dropdown by it's label and removes it if it is a multiselect.
     */
    multiRemoveByLabel: multiRemoveByLabelFn
  }
}

type selectByLabelFn = (label: string) => Cypress.Chainable
type findByLabelFn = (label: string) => Cypress.Chainable
type multiRemoveByLabelFn = (label: string) => Cypress.Chainable

function selectByLabel(sub: any, label: string): Cypress.Chainable {
  return findByLabel(sub, label).click()
}

function findByLabel(sub: any, label: string): Cypress.Chainable {
  return cy.wrap(sub).then(el => {
    const isSearchSelect =
      el.parents('[data-cy=material-select]').data('cy') === 'material-select'
    if (isSearchSelect) {
      cy.wrap(sub)
        .parents('[data-cy=material-select]')
        .should('have.attr', 'data-cy-ready', 'true')
        .find('[data-cy=search-select-input]')
        .children()
        .last() // skip the chips
        .children()
        .last() // ignore the clear button
        .find('svg') // drop-down icon
        .should('have.length', 1)
        .should('be.visible')
        .click()
        .should('not.have.focus')

      cy.focused()
        .should('be.visible')
        .type(label)

      cy.get('div[data-cy=select-dropdown]').should('not.contain', 'Loading')

      return cy
        .get('div[data-cy=select-dropdown]')
        .contains('[role=menuitem]', label)
    }

    cy.wrap(sub)
      .parent()
      .find('[role=button]')
      .click()

    return cy.get('ul[role=listbox]').contains('li', label)
  })
}

function multiRemoveByLabel(sub: any, label: string): Cypress.Chainable {
  return cy.wrap(sub).then(el => {
    const isSearchSelect =
      el.parents('[data-cy=material-select]').data('cy') === 'material-select'

    // must be a multi search select
    if (!isSearchSelect) return

    cy.wrap(sub)
      .parents('[data-cy=material-select]')
      .contains('[data-cy=multi-value]', label)
      .find('svg')
      .click()
  })
}

Cypress.Commands.add('selectByLabel', { prevSubject: 'element' }, selectByLabel)
Cypress.Commands.add('findByLabel', { prevSubject: 'element' }, findByLabel)
Cypress.Commands.add(
  'multiRemoveByLabel',
  { prevSubject: 'element' },
  multiRemoveByLabel,
)
