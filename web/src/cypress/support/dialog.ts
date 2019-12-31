declare namespace Cypress {
  interface Chainable<Subject> {
    /** Click a dialog button with the given text and wait for it to dissapear. */
    dialogFinish: typeof dialogFinish

    /** Click a dialog button with the given text. */
    dialogClick: typeof dialogClick

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

    // material Select
    if (el.attr('type') === 'hidden') {
      return cy.get(selector).selectByLabel(value)
    }

    return cy
      .get(selector)
      .clear()
      .type(value)
  })
}

function dialogForm(values: { [key: string]: string | string[] | null }): void {
  dialog()
  for (let key in values) {
    const val = values[key]
    if (val === null) continue
    fillFormField(key, val)
  }
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
        .should('not.exist')
    })
}

Cypress.Commands.add('dialogFinish', dialogFinish)
Cypress.Commands.add('dialogTitle', dialogTitle)
Cypress.Commands.add('dialogForm', dialogForm)
Cypress.Commands.add('dialogContains', dialogContains)
Cypress.Commands.add('dialogClick', dialogClick)
