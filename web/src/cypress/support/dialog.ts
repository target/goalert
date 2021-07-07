import { DateTime } from 'luxon'

function dialog(): Cypress.Chainable {
  return cy.get('[role=dialog]').should('have.length', 1).should('be.visible')
}

function dialogForm(
  values: {
    [key: string]: string | string[] | null | boolean | DateTime
  },
  parentSelector?: string,
): void {
  dialog()

  const dialogSelector = '[role=dialog] #dialog-form'
  let selector = dialogSelector
  if (parentSelector) selector = `${dialogSelector} ${parentSelector}`

  cy.form(values, selector)
}

function dialogTitle(title: string): Cypress.Chainable {
  return dialog().find('[data-cy=dialog-title]').should('contain', title)
}
function dialogContains(content: string): Cypress.Chainable {
  return dialog().should('contain', content)
}

function dialogClick(s: string): Cypress.Chainable {
  return dialog().contains('button', s).click()
}

function dialogFinish(s: string): Cypress.Chainable {
  return dialogClick(s)
    .get(`[role=dialog]`)
    .should('not.exist', { timeout: 15000 })
}

Cypress.Commands.add('dialogFinish', dialogFinish)
Cypress.Commands.add('dialogTitle', dialogTitle)
Cypress.Commands.add('dialogForm', dialogForm)
Cypress.Commands.add('dialogContains', dialogContains)
Cypress.Commands.add('dialogClick', dialogClick)
Cypress.Commands.add('dialog', dialog)
