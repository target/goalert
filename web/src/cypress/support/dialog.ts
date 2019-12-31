declare namespace Cypress {
  interface Chainable<Subject> {
    /** Click a dialog button with the given text and wait for it to dissapear. */
    dialogFinish: typeof dialogFinish

    /** Assert a dialog is present with the given title string. */
    dialogTitle: typeof dialogTitle

    /** Assert a dialog with the given content is present. */
    dialogContains: typeof dialogContains

    /** Update a dialog's form fields with the given values. */
    dialogForm: typeof dialogForm
  }
}

function fillFormField(name: string, value: string | string[]) {
  const selector = `[role=dialog] #dialog-form input[name=${name}],textarea[name=${name}]`
  return cy.get(selector).then(el => {
    const isSelect =
      el.parents('[data-cy=material-select]').data('cy') === 'material-select'
    if (isSelect) {
      if (Array.isArray(value)) {
        value.forEach(val => cy.get(selector).selectByLabel(val))
        return
      }
      return cy.get(selector).selectByLabel(value)
    }

    if (Array.isArray(value)) {
      throw new TypeError('arrays only supported for search-select inputs')
    }

    return cy
      .get(selector)
      .clear()
      .type(value)
  })
}

function dialogForm(values: { [key: string]: string | string[] }): void {
  for (let key in values) {
    fillFormField(key, values[key])
  }
}

function dialogTitle(title: string): Cypress.Chainable {
  return cy
    .get('[role=dialog] [data-cy=dialog-title]')
    .contains(title)
    .should('be.visible')
}
function dialogContains(content: string): Cypress.Chainable {
  return cy
    .get('[role=dialog]')
    .contains(content)
    .should('be.visible')
}

function dialogFinish(s: string): Cypress.Chainable {
  return cy
    .get('[role=dialog]')
    .should('be.visible')
    .contains('button', s)
    .click()
    .get('[role=dialog]')
    .should('not.exist')
}

Cypress.Commands.add('dialogFinish', dialogFinish)
Cypress.Commands.add('dialogTitle', dialogTitle)
Cypress.Commands.add('dialogForm', dialogForm)
Cypress.Commands.add('dialogContains', dialogContains)
