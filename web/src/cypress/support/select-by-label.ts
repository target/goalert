declare global {
  namespace Cypress {
    interface Chainable {
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
}

type selectByLabelFn = (label: string) => Cypress.Chainable<JQuery<HTMLElement>>
type findByLabelFn = (label: string) => Cypress.Chainable<JQuery<HTMLElement>>
type multiRemoveByLabelFn = (
  label: string,
) => Cypress.Chainable<JQuery<HTMLElement>>

function isSearchSelect(sub: JQuery<HTMLElement>): Cypress.Chainable<boolean> {
  return cy.wrap(sub).then((el) => {
    return (
      el.parents('[data-cy=material-select]').data('cy') === 'material-select'
    )
  })
}

function clearSelect(
  sub: JQuery<HTMLElement>,
): Cypress.Chainable<JQuery<HTMLElement>> {
  return cy
    .wrap(sub)
    .parents('[data-cy=material-select]')
    .should('have.attr', 'data-cy-ready', 'true')
    .find('[data-cy=search-select-input]')
    .find('button[aria-label="Clear"]')
    .should('be.visible')
    .click()
}

function findByLabel(
  sub: JQuery<HTMLElement>,
  label: string,
): Cypress.Chainable {
  return isSearchSelect(sub).then((isSearchSelect) => {
    if (isSearchSelect) {
      cy.wrap(sub)
        .parents('[data-cy=material-select]')
        .should('have.attr', 'data-cy-ready', 'true')
        .find('[data-cy=search-select-input]')
        .find('button[aria-label="Open"]')
        .should('be.visible')
        .click()

      cy.focused().should('be.visible').type(label)

      cy.get('[data-cy=select-dropdown]').should('not.contain', 'Loading')

      return cy
        .get('body')
        .contains('[data-cy=select-dropdown] [role=option]', label)
    }

    cy.wrap(sub).parent().find('[role=button]').click()

    return cy.get('ul[role=listbox]').contains('li', label)
  })
}

function selectByLabel(
  sub: JQuery<HTMLElement>,
  label: string,
): Cypress.Chainable<JQuery<HTMLElement>> {
  return isSearchSelect(sub).then((isSearchSelect) => {
    // clear value in search select
    if ((!label || label === '{backspace}') && isSearchSelect) {
      return clearSelect(sub)
    }

    return findByLabel(sub, label)
      .click()
      .get('[data-cy=select-dropdown]')
      .should('not.exist')
      .get('ul[role=listbox]')
      .should('not.exist')
  })
}

function multiRemoveByLabel(
  sub: JQuery<HTMLElement>,
  label: string,
): Cypress.Chainable {
  return isSearchSelect(sub).then((isSearchSelect) => {
    // must be a multi search select
    if (!isSearchSelect) return cy.wrap(sub)

    return cy
      .wrap(sub)
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

export {}
