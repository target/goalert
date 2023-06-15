import { Chance } from 'chance'
import { testScreen } from '../support/e2e'
import profile from '../fixtures/profile.json'
const c = new Chance()

function countryCodeCheck(
  screen: ScreenFormat,
  country: string,
  countryCode: string,
  value: string,
  formattedValue: string,
): void {
  it(`should handle ${country} phone number`, () => {
    const name = 'CM SM ' + c.word({ length: 8 })
    const type = c.pickone(['SMS', 'VOICE'])

    if (screen === 'mobile') {
      cy.pageFab('Create Contact Method')
    } else {
      cy.get('button[title="Create Contact Method"]').click()
    }

    cy.get('input[name=name]').type(name)
    cy.get('input[name=type]').selectByLabel(type)
    cy.get('input[name=value]').type(countryCode + value)
    cy.dialogFinish('Submit')
    cy.get('body').should('contain', formattedValue)
  })
}

function testProfile(screen: ScreenFormat): void {
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
          .engineTrigger()
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
      cy.url().should('eq', Cypress.config().baseUrl + '/users/' + profile.id)
    })

    it('should change the theme mode and color', () => {
      cy.get('[aria-label="Manage Profile"]').click()

      // test changing theme color
      let appbarColor: string
      cy.get('[data-cy="manage-profile"] button').contains('Light').click()
      cy.get('[data-cy="app-bar"]').then(
        (el) => (appbarColor = el.css('background-color')),
      )

      // set input of color
      cy.get(
        '[data-cy="manage-profile"] button[aria-label="More Options"]',
      ).click()
      cy.get('input[id="custom-color-picker"]')
        .invoke('val', '#fff000')
        .trigger('input')

      // assert primary color has changed
      cy.reload()
      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy="app-bar"]').then((el) =>
        expect(appbarColor).not.to.equal(el.css('background-color')),
      )

      // test changing theme mode to dark
      cy.get('[data-cy="manage-profile"] button').contains('Dark').click()

      // assert theme mode has changed
      cy.reload()
      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy="app-bar"]').then((el) =>
        expect(appbarColor).not.to.equal(el.css('background-color')),
      )
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
      if (screen === 'mobile') {
        cy.pageFab('Create Contact Method')
      } else {
        cy.get('button[title="Create Contact Method"]').click()
      }

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
          if (screen === 'mobile') {
            cy.pageFab('Create Contact Method')
          } else {
            cy.get('button[title="Create Contact Method"]').click()
          }

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

      if (screen === 'mobile') {
        cy.get('button[data-cy=page-fab]').should('be.visible').click()
        cy.get(
          `span[aria-label*=${JSON.stringify(
            'Add Notification Rule',
          )}] button[role=menuitem]`,
        ).should('be.disabled')
      } else {
        cy.get('button[title="Add Notification Rule"]').should('be.disabled')
      }
    })

    it('should display notification disclaimer when enabled', () => {
      const disclaimer = c.sentence()
      cy.updateConfig({
        General: {
          NotificationDisclaimer: disclaimer,
        },
      })
      cy.reload()

      if (screen === 'mobile') {
        cy.pageFab('Create Contact Method')
      } else {
        cy.get('button[title="Create Contact Method"]').click()
      }

      cy.dialogTitle('New Contact Method')
      cy.dialogContains(disclaimer)

      cy.updateConfig({
        General: {
          NotificationDisclaimer: 'new disclaimer',
        },
      })
      cy.reload()

      if (screen === 'mobile') {
        cy.pageFab('Create Contact Method')
      } else {
        cy.get('button[title="Create Contact Method"]').click()
      }

      cy.dialogTitle('New Contact Method')
      cy.dialogContains('new disclaimer')
    })

    countryCodeCheck(screen, 'India', '+91', '1234567890', '+91 1234 567 890')
    countryCodeCheck(screen, 'UK', '+44', '7911123456', '+44 7911 123456')

    it('should not allow fake country codes', () => {
      const value = '810' + c.integer({ min: 3000000, max: 3999999 })
      const name = 'CM SM ' + c.word({ length: 8 })
      const type = c.pickone(['SMS', 'VOICE'])

      if (screen === 'mobile') {
        cy.pageFab('Create Contact Method')
      } else {
        cy.get('button[title="Create Contact Method"]').click()
      }

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

      if (screen === 'mobile') {
        cy.pageFab('Add Notification Rule')
      } else {
        cy.get('button[title="Add Notification Rule"]').click()
      }

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

      if (screen === 'mobile') {
        cy.pageFab('Add Notification Rule')
      } else {
        cy.get('button[title="Add Notification Rule"]').click()
      }

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
