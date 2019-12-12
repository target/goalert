import { Chance } from 'chance'
import { testScreen } from '../support'
const c = new Chance()

testScreen('Profile', testProfile)

function testProfile(screen: ScreenFormat) {
  let nr: NotificationRule
  let cm: ContactMethod

  beforeEach(() =>
    cy
      .resetProfile()
      .addNotificationRule()
      .then(rule => {
        nr = rule
        cm = rule.cm
        return cy.visit('/profile')
      }),
  )

  it('should allow configuring status updates', () => {
    cy.get('input[name=alert-status-contact-method]').selectByLabel(cm.name)
    cy.get('input[name=alert-status-contact-method]').should(
      'have.value',
      cm.id,
    )
    cy.get('input[name=alert-status-contact-method]').selectByLabel('Disable')
    cy.get('input[name=alert-status-contact-method]').should(
      'not.have.value',
      cm.id,
    )
  })

  describe('Contact Methods', () => {
    it('should allow creating', () => {
      const value = '763' + c.integer({ min: 3000000, max: 3999999 })
      const name = 'SM CM ' + c.word({ length: 8 })
      const type = c.pickone(['SMS', 'VOICE'])
      const usCountryCode = '+1'

      cy.pageFab('Contact')
      cy.get('div[role=dialog]').as('dialog')

      cy.get('@dialog')
        .find('input[name=name]')
        .type(name)
      cy.get('@dialog')
        .find('input[name=type]')
        .selectByLabel(type)
      cy.get('@dialog')
        .find('input[name=value]')
        .type(usCountryCode + value)
      cy.get('@dialog')
        .find('button[type=submit]')
        .click()

      // todo: closing form pending twilio mock server verification
      cy.get(`[data-cy='verify-form']`)
        .contains('button[type=button]', 'Cancel')
        .click()

      cy.get('ul[data-cy="contact-methods"]')
        .contains('li', `${name} (${type})`)
        .find(`button[data-cy='cm-disabled']`)

      cy.get('body').should('contain', `${name} (${type})`)
    })

    it('should allow editing', () => {
      const name = 'SM CM ' + c.word({ length: 8 })
      cy.get('ul[data-cy=contact-methods]')
        .contains('li', cm.name)
        .find('button[data-cy=other-actions]')
        .menu('Edit')

      cy.get('input[name=name]')
        .clear()
        .type(name)
      cy.get('button[type=submit]').click()

      cy.get('ul[data-cy=contact-methods]').should(
        'contain',
        `${name} (${cm.type})`,
      )
    })

    it('should allow deleting', () => {
      cy.get('ul[data-cy=contact-methods]')
        .contains('li', cm.name)
        .find('button[data-cy=other-actions]')
        .menu('Delete')
      cy.get('*[role=dialog]')
        .contains('button', 'Confirm')
        .click()
      cy.get('body').should('not.contain', cm.name)
      cy.get('body').should('contain', 'No contact methods')
      cy.get('body').should('contain', 'No notification rules')
    })

    it('should disable Add Notification Rule if there are no Contact Methods', () => {
      cy.get('ul[data-cy=contact-methods]')
        .contains('li', cm.name)
        .find('button[data-cy=other-actions]')
        .menu('Delete')
      cy.get('*[role=dialog]')
        .contains('button', 'Confirm')
        .click()
      cy.get('button[data-cy=page-fab]')
        .should('be.visible')
        .click()
      cy.get(
        `span[aria-label*=${JSON.stringify(
          'Add Notification Rule',
        )}] button[role=menuitem]`,
      ).should('be.disabled')
    })

    it('should display notification disclaimer when enabled', () => {
      const disclaimer = c.sentence()
      cy.updateConfig({
        General: {
          NotificationDisclaimer: disclaimer,
        },
      })
      cy.reload()

      cy.pageFab('Add Contact Method')
      cy.get('div[role=dialog]').as('dialog')
      cy.get('@dialog')
        .find('span')
        .should('contain', disclaimer)

      cy.updateConfig({
        General: {
          NotificationDisclaimer: '',
        },
      })
      cy.reload()

      cy.pageFab('Add Contact Method')
      cy.get('div[role=dialog]').as('dialog')
      cy.get('@dialog')
        .find('span')
        .should('not.contain', disclaimer)
    })

    countryCodeCheck('India', '+91', '1234567890', '+91 1234 567 890')
    countryCodeCheck('UK', '+44', '7911123456', '+44 7911 123456')

    it('should not allow fake country codes', () => {
      const value = '810' + c.integer({ min: 3000000, max: 3999999 })
      const name = 'CM SM ' + c.word({ length: 8 })
      const type = c.pickone(['SMS', 'VOICE'])
      const fakeCountryCode = '+555'

      cy.pageFab('Contact')
      cy.get('input[name=name]').type(name)
      cy.get('input[name=type]').selectByLabel(type)
      cy.get('input[name=value]').type(fakeCountryCode + value)
      cy.get('button[type=submit]').click()
      cy.get('[aria-labelledby=countryCodeIndicator]')
        .siblings()
        .contains('Must be a valid number')
    })

    it('should set and verify contact method on first login', () => {
      cy.visit(`/alerts?isFirstLogin=1`)

      const value = '763' + c.integer({ min: 3000000, max: 3999999 })
      const name = 'SM CM ' + c.word({ length: 8 })
      const usCountryCode = '+1'

      cy.get('body').should('contain', 'Welcome to GoAlert')

      cy.get('div[role=dialog]').as('dialog')

      cy.get('@dialog')
        .find('input[name=name]')
        .type(name)
      cy.get('@dialog')
        .find('input[name=value]')
        .type(usCountryCode + value)
      cy.get('@dialog')
        .find('button[type=submit]')
        .click()

      cy.get(`[data-cy='verify-form']`)
        .contains('button[type=button]', 'Cancel')
        .click()
    })
  })

  describe('Notification Rules', () => {
    it('should allow creating an immediate rule', () => {
      // delete existing notification rule
      cy.get('ul[data-cy=notification-rules]')
        .contains('li', cm.name)
        .find('button')
        .click()
      cy.get('*[role=dialog]')
        .contains('button', 'Confirm')
        .click()

      cy.pageFab('Notification')
      cy.get('input[name=contactMethodID]').selectByLabel(cm.name)
      cy.get('*[role=dialog]')
        .contains('button', 'Submit')
        .click()
      cy.get('ul[data-cy=notification-rules]').should(
        'not.contain',
        'No notification rules',
      )
      cy.get('ul[data-cy=notification-rules]').should(
        'contain',
        `Immediately notify me via ${cm.type}`,
      )
    })

    it('should allow creating a delayed rule', () => {
      cy.get('ul[data-cy=notification-rules]')
        .contains('li', cm.name)
        .find('button')
        .click()
      cy.get('*[role=dialog]')
        .contains('button', 'Confirm')
        .click()

      const delay = c.integer({ min: 2, max: 15 })
      cy.pageFab('Notification')
      cy.get('input[name=delayMinutes]').type(delay.toString())
      cy.get('input[name=contactMethodID]').selectByLabel(cm.name)
      cy.get('*[role=dialog]')
        .contains('button', 'Submit')
        .click()

      cy.get('body').should('not.contain', 'No notification rules')
      cy.get('ul[data-cy=notification-rules]').should(
        'contain',
        `After ${delay} minutes notify me via ${cm.type}`,
      )
    })

    it('should allow deleting', () => {
      cy.get('ul[data-cy=notification-rules]')
        .contains('li', cm.name)
        .find('button')
        .click()
      cy.get('*[role=dialog]')
        .contains('button', 'Confirm')
        .click()

      cy.get('body').should('contain', 'No notification rules')
    })
  })
}

function countryCodeCheck(
  country: string,
  countryCode: string,
  value: string,
  formattedValue: string,
) {
  it(`should handle ${country} phone number`, () => {
    const name = 'CM SM ' + c.word({ length: 8 })
    const type = c.pickone(['SMS', 'VOICE'])

    cy.pageFab('Contact')
    cy.get('input[name=name]').type(name)
    cy.get('input[name=type]').selectByLabel(type)
    cy.get('input[name=value]').type(countryCode + value)
    cy.get('button[type=submit]').click()
    cy.get('body').should('contain', formattedValue)
  })
}
