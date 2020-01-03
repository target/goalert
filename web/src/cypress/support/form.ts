declare namespace Cypress {
  interface Chainable<Subject> {
    /** Update form fields with the given values. */
    form: typeof form
  }
}

function fillFormField(
  selPrefix: string,
  name: string,
  value: string | string[] | boolean,
) {
  const selector = `${selPrefix} input[name="${name}"],textarea[name="${name}"]`

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

function form(
  values: {
    [key: string]: string | string[] | null | boolean
  },
  selectorPrefix = '',
): void {
  for (let key in values) {
    const val = values[key]
    if (val === null) continue
    fillFormField(selectorPrefix, key, val)
  }
}

Cypress.Commands.add('form', form)
