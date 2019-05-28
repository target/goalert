import { Chance } from 'chance'
const c = new Chance()
import { testScreen } from '../support'

testScreen('Admin', testAdmin, false, true)

function testAdmin(screen: ScreenFormat) {
  let cfg: Config
  describe('Admin Page', () => {
    beforeEach(() => {
      return cy
        .resetConfig()
        .updateConfig({
          Mailgun: {
            APIKey: 'key-' + c.string({ length: 32, pool: '0123456789abcdef' }),
            EmailDomain: '',
          },
          Twilio: {
            AccountSID:
              'AC' + c.string({ length: 32, pool: '0123456789abcdef' }),
            AuthToken: c.string({ length: 32, pool: '0123456789abcdef' }),
            FromNumber: '+17633' + c.string({ length: 6, pool: '0123456789' }),
          },
        })
        .then(curCfg => {
          cfg = curCfg
          return cy.visit('/admin')
        })
    })

    it('should update a config value', () => {
      const newVal = 'http://' + c.domain()
      cy.get('input[name="General.PublicURL"]')
        .clear()
        .type(newVal)
      cy.get('button[data-cy="save"]').click()
      cy.get('p[data-cy="old"]').should('contain', cfg.General.PublicURL)
      cy.get('p[data-cy="new"]').should('contain', newVal)
      cy.get('button[type="submit"]')
        .contains('Confirm')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close
      cy.get('input[name="General.PublicURL"]').should('have.value', newVal)
    })

    it('should set a config value', () => {
      const newVal = c.domain()
      cy.get('input[name="Mailgun.EmailDomain"]').type(newVal)
      cy.get('button[data-cy="save"]').click()
      cy.get('p[data-cy="old"]').should('not.exist')
      cy.get('p[data-cy="new"]').should('contain', newVal)
      cy.get('button[type="submit"]')
        .contains('Confirm')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close
      cy.get('input[name="Mailgun.EmailDomain"]').should('have.value', newVal)
    })

    it('should update multiple config values at once', () => {
      const newVal1 = 'http://' + c.domain()
      const newVal2 = c.domain()

      cy.get('input[name="General.PublicURL"]')
        .clear()
        .type(newVal1)
      cy.get('input[name="Mailgun.EmailDomain"]').type(newVal2)
      cy.get('button[data-cy="save"]').click()

      cy.get('li')
        .contains('General.PublicURL')
        .siblings('p[data-cy="old"]')
        .should('contain', cfg.General.PublicURL)
      cy.get('li')
        .contains('General.PublicURL')
        .siblings('p[data-cy="new"]')
        .should('contain', newVal1)

      cy.get('li')
        .contains('Mailgun.EmailDomain')
        .siblings('p[data-cy="old"]')
        .should('not.exist') // not set in beforeEach
      cy.get('li')
        .contains('Mailgun.EmailDomain')
        .siblings('p[data-cy="new"]')
        .should('contain', newVal2)

      cy.get('button[type="submit"]')
        .contains('Confirm')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close

      cy.get('input[name="General.PublicURL"]').should('have.value', newVal1)
      cy.get('input[name="Mailgun.EmailDomain"]').should('have.value', newVal2)
    })

    it('should delete a config value', () => {
      cy.get('input[name="General.PublicURL"]').clear()
      cy.get('button[data-cy="save"]').click()
      cy.get('p[data-cy="old"]').should('contain', cfg.General.PublicURL)
      cy.get('p[data-cy="new"]').should('not.exist')
      cy.get('button[type="submit"]')
        .contains('Confirm')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close
      cy.get('input[name="General.PublicURL"]').should('have.value', '')
    })

    it('should cancel changing a config value', () => {
      const newVal = c.domain()
      cy.get('input[name="General.PublicURL"]')
        .clear()
        .type(newVal)
      cy.get('button[data-cy="save"]').click()
      cy.get('p[data-cy="old"]').should('contain', cfg.General.PublicURL)
      cy.get('p[data-cy="new"]').should('contain', newVal)
      cy.get('button[type="button"]')
        .contains('Cancel')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close
    })

    it('should reset pending config value changes', () => {
      cy.get('input[name="General.PublicURL"]')
        .clear()
        .type(c.domain())
      cy.get('input[name="Mailgun.EmailDomain"]').type(c.domain())

      cy.get('button[data-cy="reset"]').click()

      cy.get('input[name="General.PublicURL"]').should(
        'have.value',
        cfg.General.PublicURL,
      )
      cy.get('input[name="Mailgun.EmailDomain"]').should('have.value', '')
    })

    it('should update a boolean toggle field', () => {
      cy.get('input[name="Twilio.Enable"]').click()
      cy.get('button[data-cy="save"]').click()
      cy.get('p[data-cy="old"]').should('contain', 'false')
      cy.get('p[data-cy="new"]').should('contain', 'true')
      cy.get('button[type="submit"]')
        .contains('Confirm')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close
      cy.get('input[name="Twilio.Enable"]').should('have.value', 'true')
    })

    it('should update a string list field', () => {
      const domain1 = 'http://' + c.domain()
      const domain2 = 'http://' + c.domain()
      const domain3 = cfg.General.PublicURL

      cy.get('input[name="Auth.RefererURLs-0"]')
        .clear()
        .type(domain1)
      cy.get('input[name="Auth.RefererURLs-new-item"]')
        .clear()
        .type(domain2)
      cy.get('input[name="Auth.RefererURLs-new-item"]')
        .clear()
        .type(domain3)

      cy.get('button[data-cy="save"]').click()
      // cy.get('p[data-cy="old"]').should('contain', url)
      cy.get('p[data-cy="new"]').should(
        'contain',
        domain1 + ', ' + domain2 + ', ' + domain3,
      )
      cy.get('button[type="submit"]')
        .contains('Confirm')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close
      cy.get('input[name="Auth.RefererURLs-0"]').should('have.value', domain1)
      cy.get('input[name="Auth.RefererURLs-1"]').should('have.value', domain2)
      cy.get('input[name="Auth.RefererURLs-2"]').should('have.value', domain3)
      cy.get('input[name="Auth.RefererURLs-new-item"]').should('have.value', '')
    })

    it('should update a password field', () => {
      const newKey = 'key-' + c.string({ length: 32, pool: '0123456789abcdef' })
      cy.get('input[name="Mailgun.APIKey"]')
        .clear()
        .type(newKey)
      cy.get('button[data-cy="save"]').click()
      cy.get('p[data-cy="old"]').should('contain', cfg.Mailgun.APIKey)
      cy.get('p[data-cy="new"]').should('contain', newKey)
      cy.get('button[type="submit"]')
        .contains('Confirm')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close
      cy.get('input[name="Mailgun.APIKey"]').should('have.value', newKey)
    })
  })
}
