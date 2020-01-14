declare namespace Cypress {
  interface Chainable<Subject> {
    /** Update form fields with the given values. */
    form: typeof form
  }
}

function getSelector(selPrefix: string, name: string) {
  return `${selPrefix} input[name="${name}"],textarea[name="${name}"]`
}

function fillFormField(
  selPrefix: string,
  name: string,
  value: string | string[] | boolean,
) {
  const selector = getSelector(selPrefix, name)

  if (typeof value === 'boolean') {
    if (!value) return cy.get(selector).uncheck()

    return cy.get(selector).check()
  }

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

    if (value === '') return cy.get(selector).clear()

    return cy
      .get(selector)
      .clear()
      .type(value)
  })
}

/* clearFormField
 * Clears select input, text input fields
 * Does not handle checkboxes; use fillFormField(...args, false) instead
 * Does not handle radio buttons
 * Does not handle hidden input
 */
function clearFormField(selPrefix: string, name: string) {
  const selector = getSelector(selPrefix, name)

  return cy.get(selector).then(el => {
    const isSelect =
      el.parents('[data-cy=material-select]').data('cy') === 'material-select'

    if (isSelect) {
      const searchSelectInput = `[role=dialog] #dialog-form [data-cy=search-select-input][name="${name}"] input`
      cy.get(searchSelectInput)
        .click()
        .clear()

      //click label to deselect
      cy.get('[data-cy=material-select] label').click()
      return
    }

    return cy.get(selector).clear()
  })
}

function form(
  values: {
    [key: string]: string | string[] | null | boolean
  },
  selectorPrefix = '',
): void {
  for (let key in values) {
    const val = values[key]
    if (val === null) {
      clearFormField(selectorPrefix, key)
    } else {
      fillFormField(selectorPrefix, key, val)
    }
  }
}

Cypress.Commands.add('form', form)
