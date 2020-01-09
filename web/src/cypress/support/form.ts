import { DateTime } from 'luxon'
declare global {
  namespace Cypress {
    interface Chainable {
      /** Update form fields with the given values. */
      form: typeof form
    }
  }
}

const clickArc = (pct: number) => (el: any) => {
  const height = el.height() || 0
  const radius = height / 2
  const angle = (-1 + pct) * Math.PI * 2 - Math.PI / 2

  const x = radius * Math.cos(angle) * 0.9 + radius
  const y = radius * Math.sin(angle) * 0.9 + radius

  return cy.wrap(el).click(x, y)
}

// materialClock will control a material time-picker from an input field
function materialClock(selector: string, time: string) {
  const parts = time.split(':')
  let hour = +parts[0]
  const min = +parts[1]

  const isAM = hour < 12
  if (!isAM) hour -= 12

  return cy
    .get(selector)
    .click() // open dialog

    .get('[role=dialog][data-cy=picker-fallback]')
    .should('be.visible')

    .contains('button', isAM ? 'AM' : 'PM')
    .click() // select AM or PM

    .get('[role=dialog][data-cy=picker-fallback] [role=menu]')
    .parent()
    .then(clickArc(hour / 12)) // select the hour

    .get('[role=dialog][data-cy=picker-fallback] [role=menu]')
    .parent()
    .then(clickArc(min / 60)) // minutes

    .get('[role=dialog][data-cy=picker-fallback]')
    .should('not.exist') // wait for dialog to dissapear
}

// materialCalendar will control a material date-picker from an input field
function materialCalendar(selector: string, date: string) {
  const dt = DateTime.fromFormat(date, 'yyyy-MM-dd')

  cy.get(selector).click() // open dialog
  cy.get('[role=dialog][data-cy=picker-fallback]')
    .should('be.visible')

    .find('button')
    .first()
    .click() // open year selection

  cy.get('[role=dialog][data-cy=picker-fallback]')
    .contains('[role=button]', dt.year)
    .click() // click on correct year

  cy.get(
    '[role=dialog][data-cy=picker-fallback] button[data-cy=month-back]+div',
  ).then(el => {
    const displayedDT = DateTime.fromFormat(el.text(), 'MMMM yyyy')
    const diff = dt.startOf('month').diff(displayedDT, 'months').months

    // navigate to correct month
    for (let i = 0; i < Math.abs(diff); i++) {
      cy.get(`button[data-cy=month-${diff < 0 ? 'back' : 'next'}]`).click()

      cy.get(
        '[role=dialog][data-cy=picker-fallback] button[data-cy=month-back]+div',
      )

        .should(
          'contain',
          displayedDT
            .plus({ months: (diff < 0 ? -1 : 1) * i + 1 })
            .toFormat('MMMM'),
        )
        .should(
          'not.contain',
          displayedDT
            .plus({ months: (diff < 0 ? -1 : 1) * i })
            .toFormat('MMMM'),
        )
    }

    cy.wait(3000)

    // click on the day
    cy.get('body')
      .contains('button', dt.day)
      .last()
      .should('have.length', 1)
      .click({ force: true })

    // wait for dialog to dissapear
    cy.get('[role=dialog][data-cy=picker-fallback]').should('not.exist')
  })
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
    const pickerFallback = el
      .parents('[data-cy-fallback-type]')
      .data('cyFallbackType')
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

    if (value === '') return cy.get(selector).clear()

    // material Select
    if (el.attr('type') === 'hidden')
      return cy.get(selector).selectByLabel(value)

    if (pickerFallback) {
      switch (pickerFallback) {
        case 'time':
          return materialClock(selector, value)
        case 'date':
          return materialCalendar(selector, value)
      }
    }

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
