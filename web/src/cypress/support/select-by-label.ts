function isSearchSelect(sub: HTMLElement): Cypress.Chainable<boolean> {
  return cy.wrap(sub).then((el) => {
    return (
      el.parents('[data-cy=material-select]').data('cy') === 'material-select'
    )
  })
}

function clearSelect(sub: HTMLElement): Cypress.Chainable<JQuery<HTMLElement>> {
  return cy
    .wrap(sub)
    .parents('[data-cy=material-select]')
    .should('have.attr', 'data-cy-ready', 'true')
    .find('[data-cy=search-select-input]')
    .children()
    .last() // skip the chips
    .children()
    .first() // get the clear button
    .find('svg') // clear field icon
    .should('have.length', 1)
    .should('be.visible')
    .click()
}

function findByLabel(
  sub: HTMLElement,
  label: string,
): Cypress.Chainable<JQuery<HTMLElement>> {
  return isSearchSelect(sub).then((isSearchSelect) => {
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

      cy.focused().should('be.visible').type(label)

      cy.get('[data-cy=select-dropdown]').should('not.contain', 'Loading')

      return cy.get('[data-cy=select-dropdown] [role=menuitem]').contains(label)
    }

    cy.wrap(sub).parent().find('[role=button]').click()

    return cy.get('ul[role=listbox]').contains('li', label)
  })
}

function selectByLabel(
  sub: HTMLElement,
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
  sub: HTMLElement,
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
