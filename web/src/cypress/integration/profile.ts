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

      cy.pageFab('Contact')
      cy.get(`[data-cy='create-form']`)
        .get('input[name=name]')
        .type(name)
        .get('input[name=type]')
        .selectByLabel(type)
        .get('input[name=value]')
        .type(value)
        .get('button[type=submit]')
        .click()
      cy.get(`[data-cy='verify-form']`)
        .contains('button[type=button]', 'Cancel')
        .click()
      cy.get('ul[data-cy="contact-methods"]')
        .contains('li', `${name} (${type})`)
        .find(`button[data-cy='cm-disabled']`)

      // TODO: twilio mock server verification pending

      cy.get('body').should('contain', `${name} (${type})`)
    })
    it('should allow editing', () => {
      const name = 'SM CM ' + c.word({ length: 8 })
      const value = '763' + c.integer({ min: 3000000, max: 3999999 })
      cy.get('ul[data-cy=contact-methods]')
        .contains('li', cm.name)
        .find('button[data-cy=other-actions]')
        .menu('Edit')

      cy.get('input[name=name]')
        .clear()
        .type(name)
      cy.get('input[name=value]')
        .clear()
        .type(value)
      cy.get('button[type=submit]').click()

      cy.get('ul[data-cy=contact-methods]').should(
        'contain',
        `${name} (${cm.type})`,
      )
      cy.get('ul[data-cy="contact-methods"]')
        .contains('li', `${name} (${cm.type})`)
        .find(`button[data-cy='cm-disabled']`)
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
    it('should display notification disclaimer when enabled', () => {
      const sentence = c.sentence()
      cy.updateConfig({
        General: {
          NotificationDisclaimer: sentence,
        },
      })
      cy.reload()

      cy.get('body').should('contain', sentence)

      cy.updateConfig({
        General: {
          NotificationDisclaimer: '',
        },
      })
      cy.reload()
      cy.get('body').should('not.contain', sentence)
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
