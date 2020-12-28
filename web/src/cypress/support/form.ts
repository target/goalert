import { DateTime } from 'luxon'

const clickArc = (pct: number) => (el: JQuery) => {
  const height = el.height() || 0
  const radius = height / 2
  const angle = (-1 + pct) * Math.PI * 2 - Math.PI / 2

  const x = radius * Math.cos(angle) * 0.9 + radius
  const y = radius * Math.sin(angle) * 0.9 + radius

  return cy.wrap(el).click(x, y)
}

// materialClock will control a material time-picker from an input field
function materialClock(
  time: string | DateTime,
): Cypress.Chainable<JQuery<HTMLElement>> {
  const dt = DateTime.isDateTime(time)
    ? time
    : DateTime.fromFormat(time, 'HH:mm')

  let hour = dt.hour

  const isAM = hour < 12
  if (!isAM) hour -= 12

  return cy
    .contains('button', isAM ? 'AM' : 'PM')
    .click() // select AM or PM

    .get('[role=dialog][data-cy=picker-fallback] [role=menu]')
    .parent()
    .then(clickArc(hour / 12)) // select the hour

    .get('[role=dialog][data-cy=picker-fallback] [role=menu]')
    .parent()
    .then(clickArc(dt.minute / 60)) // minutes
}

const openPicker = (selector: string): void => {
  cy.get(selector).click() // open dialog
  cy.get('[role=dialog][data-cy=picker-fallback]').should('be.visible')
}
const finishPicker = (): void => {
  cy.get('[role=dialog][data-cy=picker-fallback]')
    .contains('button', 'OK')
    .click()
  // wait for dialog to dissapear
  cy.get('[role=dialog][data-cy=picker-fallback]').should('not.exist')
}

// materialCalendar will control a material date-picker from an input field
function materialCalendar(date: string | DateTime): void {
  const dt = DateTime.isDateTime(date)
    ? date
    : DateTime.fromFormat(date, 'yyyy-MM-dd')

  cy.get('[role=dialog][data-cy=picker-fallback]')
    .find('button')
    .first()
    .click() // open year selection

  cy.get('[role=dialog][data-cy=picker-fallback]')
    .contains('[role=button]', dt.year)
    .click() // click on correct year

  cy.get(
    '[role=dialog][data-cy=picker-fallback] button[data-cy=month-back]+div',
  ).then((el) => {
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
            .plus({ months: (diff < 0 ? -1 : 1) * (i + 1) })
            .toFormat('MMMM'),
        )
        .should(
          'not.contain',
          displayedDT
            .plus({ months: (diff < 0 ? -1 : 1) * i })
            .toFormat('MMMM'),
        )
    }

    // click on the day
    cy.get('body')
      .contains("button[tabindex='0']", new RegExp(`^${dt.day.toString()}$`))
      .click({ force: true })
  })
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

      const pickerFallback = el
        .parents('[data-cy-fallback-type]')
        .data('cyFallbackType')

      if (isSelect) {
        if (value === '') {
          cy.get(selector).clear()

          // clear chips on multi-select
          el.parent()
            .find('[data-cy="multi-value"] > svg')
            .each(() => {
              cy.get(selector).type(`{backspace}`)
            })

          return cy.get(selector).clear()
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

      if (pickerFallback) {
        switch (pickerFallback) {
          case 'time':
            openPicker(selector)
            materialClock(value)
            finishPicker()
            return
          case 'date':
            openPicker(selector)
            materialCalendar(value)
            finishPicker()
            return
          case 'datetime-local':
            openPicker(selector)
            materialCalendar(value)
            materialClock(value)
            finishPicker()
            return
          default:
            if (DateTime.isDateTime(value)) {
              throw new TypeError(
                'DateTime only supported for time, date, or datetime-local types',
              )
            }
        }
      }

      return cy.get(selector).then((el) => {
        if (!DateTime.isDateTime(value)) {
          if (el.attr('type') === 'hidden') {
            return cy.get(selector).selectByLabel(value)
          }
          return cy.wrap(el).clear().type(value)
        }

        // material Select
        switch (el.attr('type')) {
          case 'time':
            return cy.wrap(el).clear().type(value.toFormat('HH:mm'))
          case 'date':
            return cy.wrap(el).clear().type(value.toFormat('yyyy-MM-dd'))
          case 'datetime-local':
            return cy
              .wrap(el)
              .clear()
              .type(value.toFormat(`yyyy-MM-dd'T'HH:mm`))
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
