import { DateTime } from 'luxon'

declare global {
  namespace Cypress {
    interface Chainable {
      /** Update form fields with the given values. */
      form: typeof form
    }
  }
}

function fillFormField(
  selPrefix: string,
  name: string,
  value: string | string[] | boolean | DateTime,
): Cypress.Chainable<JQuery<HTMLElement>> {
  const selector = `${selPrefix} input[name="${name}"],textarea[name="${name}"]`

  return cy
    .get(selector)
    .then((el) => {
      if (el.attr('type') === 'radio') {
        return cy
          .get(`${selPrefix} input[name="${name}"][value="${value}"]`)
          .click()
      }
      // Auto detect/expand accordion sections if need be
      const accordionSectionID = el
        .parents('[aria-labelledby][role=region]')
        .attr('id')
      if (accordionSectionID) {
        const ctrl = `[aria-controls=${accordionSectionID}][aria-expanded=false]`
        if (Cypress.$(ctrl).length > 0) {
          cy.get(ctrl).click()
        }
      }

      if (typeof value === 'boolean') {
        if (!value) return cy.get(selector).uncheck()

        return cy.get(selector).check()
      }

      const isSelect =
        el.parents('[data-cy=material-select]').data('cy') ===
          'material-select' ||
        el.siblings('[role=button]').attr('aria-haspopup') === 'listbox'

      if (isSelect) {
        if (value === '') {
          cy.get(selector).clear()

          // clear chips on multi-select
          el.parent()
            .find('[data-cy="multi-value"]')
            .each(() => {
              cy.get(selector).type(`{backspace}`)
            })

          return cy.get(selector)
        }

        if (DateTime.isDateTime(value)) {
          throw new TypeError(
            'DateTime only supported for time, date, or datetime-local types',
          )
        }

        if (Array.isArray(value)) {
          value.forEach((val) => cy.get(selector).selectByLabel(val))
          return
        }

        return cy.get(selector).selectByLabel(value)
      }

      if (Array.isArray(value)) {
        throw new TypeError('arrays only supported for search-select inputs')
      }

      if (value === '') return cy.get(selector).clear()

      return cy.get(selector).then((el) => {
        if (!DateTime.isDateTime(value)) {
          if (el.attr('type') === 'hidden') {
            return cy.get(selector).selectByLabel(value)
          }
          cy.wrap(el).clear()
          return cy.focused().type(value)
        }

        cy.wrap(el).clear()
        // material Select
        switch (el.attr('type')) {
          case 'time':
            return cy.focused().type(value.toFormat('HH:mm'))
          case 'date':
            return cy.focused().type(value.toFormat('yyyy-MM-dd'))
          case 'datetime-local':
            return cy.focused().type(value.toFormat(`yyyy-MM-dd'T'HH:mm`))
          default:
            throw new TypeError(
              'DateTime only supported for time, date, or datetime-local types',
            )
        }
      })
    })
    .then(() => cy.get(selector)) as Cypress.Chainable<JQuery<HTMLElement>>
}

function form(
  values: {
    [key: string]: string | string[] | null | boolean | DateTime
  },
  selectorPrefix = '',
): void {
  for (const key in values) {
    const val = values[key]
    if (val === null) continue
    fillFormField(selectorPrefix, key, val)
  }
}

Cypress.Commands.add('form', form)
