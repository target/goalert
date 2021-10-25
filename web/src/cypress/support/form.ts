import { DateTime } from 'luxon'

const clickArc = (pct: number) => (el: JQuery) => {
  const height = el.height() || 0
  const radius = height / 2
  const angle = (-1 + pct) * Math.PI * 2 - Math.PI / 2

  const x = radius * Math.cos(angle) * 0.9 + radius
  const y = radius * Math.sin(angle) * 0.9 + radius

  return cy.wrap(el).click(x, y)
}

const openPicker = (selector: string, name: string): void => {
  cy.get(selector).parent().find('button').click() // open clock/calendar popper
  cy.get(`[role=dialog][data-cy="${name}-picker-fallback"]`).should(
    'be.visible',
  )
}

// materialClock will control a material time-picker from an input field
function materialClock(
  time: string | DateTime,
  fieldName: string,
): Cypress.Chainable<JQuery<HTMLElement>> {
  const dt = DateTime.isDateTime(time)
    ? time
    : DateTime.fromFormat(time, 'HH:mm')

  let hour = dt.hour

  const isAM = hour < 12
  if (!isAM) hour -= 12

  return cy
    .get(`[role=dialog][data-cy="${fieldName}-picker-fallback"]`)
    .contains('button', isAM ? 'AM' : 'PM')
    .click() // select AM or PM

    .get(`[role=dialog][data-cy="${fieldName}-picker-fallback"] [role=listbox]`)
    .parent()
    .children()
    .eq(0)
    .then(clickArc(hour / 12)) // select the hour

    .get(`[role=dialog][data-cy="${fieldName}-picker-fallback"] [role=listbox]`)
    .parent()
    .children()
    .eq(0)
    .then(clickArc(dt.minute / 60)) // minutes
}

// materialCalendar will control a material date-picker from an input field
function materialCalendar(date: string | DateTime, fieldName: string): void {
  const dt = DateTime.isDateTime(date)
    ? date
    : DateTime.fromFormat(date, 'yyyy-MM-dd')

  cy.get(`[role=dialog][data-cy="${fieldName}-picker-fallback"]`)
    .find('button[aria-label="calendar view is open, switch to year view"]')
    .click() // open year selection

  cy.get(`[role=dialog][data-cy="${fieldName}-picker-fallback"]`)
    .contains('[type=button]', dt.year)
    .click() // click on correct year

  cy.get(
    `[role=dialog][data-cy="${fieldName}-picker-fallback"] button[aria-label="Previous month"]`,
  )
    .parent()
    .siblings()
    .then((el) => {
      const displayedDT = DateTime.fromFormat(el.text(), 'MMMMyyyy')
      const diff = dt.startOf('month').diff(displayedDT, 'months').months

      // navigate to correct month
      for (let i = 0; i < Math.abs(diff); i++) {
        cy.get(
          `button[aria-label="${diff < 0 ? 'Previous' : 'Next'} month"]`,
        ).click()

        cy.get(`[role=dialog][data-cy="${fieldName}-picker-fallback"]`)
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
      cy.get(
        `[role=dialog][data-cy="${fieldName}-picker-fallback"] button[aria-label="${dt.toFormat(
          'MMM d, y',
        )}"]`,
      ).click({ force: true })
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

      const pickerFallback = el
        .parents('[data-cy-fallback-type]')
        .data('cyFallbackType')

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

      if (pickerFallback) {
        switch (pickerFallback) {
          case 'time':
            openPicker(selector, name)
            materialClock(value, name)
            return
          case 'date':
            openPicker(selector, name)
            materialCalendar(value, name)
            return
          case 'datetime-local':
            openPicker(selector, name)
            materialCalendar(value, name)
            materialClock(value, name)
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
