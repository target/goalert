import { DateTime } from 'luxon'

function dialogForm(values: {
  [key: string]: string | string[] | null | boolean | DateTime
}): void {
  dialog()
  cy.form(values, '[role=dialog] #dialog-form')
}

function dialog(): Cypress.Chainable {
  return cy
    .get('[data-cy=unmounting]')
    .should('not.exist')
    .get('[role=dialog]')
    .should('have.length', 1)
    .should('be.visible')
}
function dialogTitle(title: string): Cypress.Chainable {
  return dialog()
    .find('[data-cy=dialog-title]')
    .should('contain', title)
}
function dialogContains(content: string): Cypress.Chainable {
  return dialog().should('contain', content)
}

function dialogClick(s: string): Cypress.Chainable {
  return dialog()
    .contains('button', s)
    .click()
}

function dialogFinish(s: string): Cypress.Chainable {
  return dialog()
    .get('[data-cy-gu]')
    .then(el => {
      const id = el.data('cyGu')
      return cy
        .dialogClick(s)
        .get(`[data-cy-gu=${id}]`)
        .should('not.exist', { timeout: 15000 })
    })
}

Cypress.Commands.add('dialogFinish', dialogFinish)
Cypress.Commands.add('dialogTitle', dialogTitle)
Cypress.Commands.add('dialogForm', dialogForm)
Cypress.Commands.add('dialogContains', dialogContains)
Cypress.Commands.add('dialogClick', dialogClick)
