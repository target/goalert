import { Chance } from 'chance'
import { testScreen } from '../support'
import profile from '../fixtures/profile.json'
const c = new Chance()

function countryCodeCheck(
  country: string,
  countryCode: string,
  value: string,
  formattedValue: string,
): void {
  it(`should handle ${country} phone number`, () => {
    const name = 'CM SM ' + c.word({ length: 8 })
    const type = c.pickone(['SMS', 'VOICE'])

    cy.pageFab('Contact')
    cy.get('input[name=name]').type(name)
    cy.get('input[name=type]').selectByLabel(type)
    cy.get('input[name=value]').type(countryCode + value)
    cy.dialogFinish('Submit')
    cy.get('body').should('contain', formattedValue)
  })
}

function testProfile(): void {
  let cm: ContactMethod

  beforeEach(() =>
    cy
      .resetProfile()
      .addNotificationRule()
      .then((rule: NotificationRule) => {
        cm = rule.contactMethod
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

  it('should list and link on-call services', () => {
    const name = 'SVC ' + c.word({ length: 8 })

    return cy
      .createService({ name })
      .then((svc: Service) => {
        return cy
          .createEPStep({
            epID: svc.epID,
            targets: [{ type: 'user', id: profile.id }],
          })
          .task('engine:trigger')
          .then(() => svc.id)
      })
      .then((svcID: string) => {
        cy.get('body').contains('a', 'On-Call').click()

        cy.get('body').contains('a', name).click()

        cy.url().should('eq', Cypress.config().baseUrl + '/services/' + svcID)
      })
  })

  describe('Settings', () => {
    it('should visit profile', () => {
      cy.visit('/')
      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy="manage-profile"]')
        .find('button')
        .contains('Manage Profile')
        .click()
      cy.url().should('eq', Cypress.config().baseUrl + '/profile')
    })

    it('should change the theme mode', () => {
      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy="manage-profile"]')
        .find('button')
        .contains('Light')
        .click()

      let lightModeColor: string, darkModeColor: string
      cy.get('div[id="app-root"]').then(
        (el) => (lightModeColor = el.css('background-color')),
      )

      cy.get('[data-cy="manage-profile"]')
        .find('button')
        .contains('Dark')
        .click()
      cy.get('div[id="app-root"]')
        .then((el) => (darkModeColor = el.css('background-color')))
        .then(() => {
          expect(lightModeColor).not.to.equal(darkModeColor)
        })
    })

    it('should not display feedback by default', () => {
      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy=feedback]').should('not.exist')
    })

    it('should display feedback with default href when enabled', () => {
      cy.updateConfig({ Feedback: { Enable: true } })
      cy.reload()
      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy="manage-profile"]')
        .find('[data-cy=feedback]')
        .should('have.attr', 'href')
        .and(
          'match',
          /https:\/\/www\.surveygizmo\.com\/s3\/4106900\/GoAlert-Feedback/,
        )
    })

    it('should display feedback with correct href when overridden', () => {
      cy.updateConfig({
        Feedback: { Enable: true, OverrideURL: 'https://www.goalert.me' },
      }).then(() => {
        cy.get('[aria-label="Manage Profile"]').click()
        cy.get('[data-cy="manage-profile"]')
          .find('[data-cy=feedback]')
          .should('have.attr', 'href')
          .and('match', /https:\/\/www\.goalert\.me/)
      })
    })
  })

  describe('Contact Methods', () => {
    function check(name: string, type: string, value: string): void {
      cy.pageFab('Contact')
      cy.dialogTitle('Create New Contact Method')
      cy.dialogForm({
        name,
        type,
        value,
      })
      cy.dialogFinish('Submit')

      // todo: closing form pending mock server verification
      cy.dialogTitle('Verify Contact Method')
      cy.dialogFinish('Cancel')

      cy.get('ul[data-cy="contact-methods"]')
        .contains('li', `${name} (${type})`)
        .find(`button[data-cy='cm-disabled']`)

      cy.get('body').should('contain', `${name} (${type})`)
    }

    it('should allow creating sms/voice', () => {
      cy.updateConfig({
        Twilio: {
          Enable: true,
          AccountSID: 'AC' + c.string({ length: 32, pool: '0123456789abcdef' }),
          AuthToken: c.string({ length: 32, pool: '0123456789abcdef' }),
          FromNumber: '+17633' + c.string({ length: 6, pool: '0123456789' }),
        },
      })
      cy.reload()

      const value = '+1763' + c.integer({ min: 3000000, max: 3999999 })
      const name = 'SM CM ' + c.word({ length: 8 })
      const type = c.pickone(['SMS', 'VOICE'])

      check(name, type, value)
    })

    it('should allow creating email', () => {
      cy.updateConfig({
        SMTP: {
          Enable: true,
        },
      })
      cy.reload()

      const name = 'SM CM ' + c.word({ length: 8 })
      const type = c.pickone(['EMAIL'])
      const value = c.email()
      check(name, type, value)
    })

    it('should allow creating webhook', () => {
      cy.updateConfig({
        Webhook: {
          Enable: true,
        },
      })
      cy.reload()

      const name = 'SM CM ' + c.word({ length: 8 })
      const type = c.pickone(['WEBHOOK'])
      const value = c.url()
      check(name, type, value)
    })

    it('should return error with link to conflicting user', () => {
      cy.addContactMethod({ userID: profile.id }).then(
        (contactMethod: ContactMethod) => {
          cy.pageFab('Add Contact Method')
          cy.dialogTitle('Create New Contact Method')
          cy.dialogForm({
            name: c.word({ length: 8 }),
            type: contactMethod.type,
            value: contactMethod.value,
          })
          cy.dialogClick('Submit')
          cy.dialog()
            .find('a[data-cy=error-help-link]')
            .should(
              'contain',
              'Contact method already exists for that type and value: ' +
                profile.name,
            )
            .should('have.attr', 'href')
            .and('include', `/users/${profile.id}`)
        },
      )
    })

    it('should allow editing', () => {
      const name = 'SM CM ' + c.word({ length: 8 })
      cy.get('ul[data-cy=contact-methods]')
        .contains('li', cm.name)
        .find('button[data-cy=other-actions]')
        .menu('Edit')

      cy.dialogTitle('Edit Contact Method')
      cy.dialogForm({ name })
      cy.dialogFinish('Submit')

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
      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.get('body').should('not.contain', cm.name)
      cy.get('body').should('contain', 'No contact methods')
      cy.get('body').should('contain', 'No notification rules')
    })

    it('should disable Add Notification Rule if there are no Contact Methods', () => {
      cy.get('ul[data-cy=contact-methods]')
        .contains('li', cm.name)
        .find('button[data-cy=other-actions]')
        .menu('Delete')

      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.get('button[data-cy=page-fab]').should('be.visible').click()
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
      cy.dialogTitle('New Contact Method')
      cy.dialogContains(disclaimer)

      cy.updateConfig({
        General: {
          NotificationDisclaimer: 'new disclaimer',
        },
      })
      cy.reload()

      cy.pageFab('Add Contact Method')
      cy.dialogTitle('New Contact Method')
      cy.dialogContains('new disclaimer')
    })

    countryCodeCheck('India', '+91', '1234567890', '+91 1234 567 890')
    countryCodeCheck('UK', '+44', '7911123456', '+44 7911 123456')

    it('should not allow fake country codes', () => {
      const value = '810' + c.integer({ min: 3000000, max: 3999999 })
      const name = 'CM SM ' + c.word({ length: 8 })
      const type = c.pickone(['SMS', 'VOICE'])

      cy.pageFab('Contact')
      cy.dialogTitle('New Contact Method')
      cy.dialogForm({
        name,
        type,
        value,
      })
      cy.dialogClick('Submit')
      cy.get('[aria-labelledby=countryCodeIndicator]')
        .siblings()
        .contains('Must be a valid number')
    })

    it('should set and verify contact method on first login', () => {
      cy.visit(`/alerts?isFirstLogin=1`)

      const value = '+1763' + c.integer({ min: 3000000, max: 3999999 })
      const name = 'SM CM ' + c.word({ length: 8 })

      cy.get('body').should('contain', 'Welcome to GoAlert')

      cy.dialogTitle('Welcome to GoAlert')
      cy.dialogForm({
        name,
        value,
      })
      cy.dialogFinish('Submit')

      cy.dialogTitle('Verify')
      cy.dialogFinish('Cancel')
    })
  })

  describe('Notification Rules', () => {
    it('should allow creating an immediate rule', () => {
      // delete existing notification rule
      cy.get('ul[data-cy=notification-rules]')
        .contains('li', cm.name)
        .find('button')
        .click()
      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.pageFab('Notification')
      cy.dialogTitle('New Notification Rule')
      cy.dialogForm({ contactMethodID: cm.name })
      cy.dialogFinish('Submit')

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
      // delete default rule
      cy.get('ul[data-cy=notification-rules]')
        .contains('li', cm.name)
        .find('button')
        .click()
      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      // create new rule
      const delay = c.integer({ min: 2, max: 15 })
      cy.pageFab('Notification')
      cy.dialogTitle('New Notification Rule')
      cy.dialogForm({
        contactMethodID: cm.name,
        delayMinutes: delay.toString(),
      })
      cy.dialogFinish('Submit')

      // verify changes
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

      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.get('body').should('contain', 'No notification rules')
    })
  })
}

testScreen('Profile', testProfile)
