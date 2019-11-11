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
          return cy
            .visit('/admin')
            .get('button[data-cy=save]')
            .should('exist')
        })
    })

    it('should update a config value', () => {
      const newVal = 'http://' + c.domain()

      cy.get('input[name="General.PublicURL"]')
        .clear()
        .should('be.empty')
        .type(newVal)
        .should('have.value', newVal)
      cy.get('button[data-cy="save"]').click()

      // save dialog
      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-General.PublicURL"] p[data-cy="old"]',
      )
        .should('contain', cfg.General.PublicURL)
        .siblings('p[data-cy="new"]')
        .should('contain', newVal)
      cy.get('button[type="submit"]')
        .contains('Confirm')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close

      cy.get('input[name="General.PublicURL"]').should('have.value', newVal)
    })

    it('should set a config value', () => {
      const newVal = c.domain()

      cy.get('input[name="Mailgun.EmailDomain"]')
        .type(newVal)
        .should('have.value', newVal)
      cy.get('button[data-cy="save"]').click()

      // save dialog
      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-Mailgun.EmailDomain"] p[data-cy="old"]',
      ).should('not.exist')
      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-Mailgun.EmailDomain"] p[data-cy="new"]',
      ).should('contain', newVal)
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
        .should('have.value', newVal1)
      cy.get('input[name="Mailgun.EmailDomain"]')
        .type(newVal2)
        .should('have.value', newVal2)
      cy.get('button[data-cy="save"]').click()

      // save dialog
      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-General.PublicURL"] p[data-cy="old"]',
      )
        .should('contain', cfg.General.PublicURL)
        .siblings('p[data-cy="new"]')
        .should('contain', newVal1)

      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-Mailgun.EmailDomain"] p[data-cy="old"]',
      ).should('not.exist') // not set in beforeEach

      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-Mailgun.EmailDomain"] p[data-cy="new"]',
      ).should('contain', newVal2)

      cy.get('button[type="submit"]')
        .contains('Confirm')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close

      cy.get('input[name="General.PublicURL"]').should('have.value', newVal1)
      cy.get('input[name="Mailgun.EmailDomain"]').should('have.value', newVal2)
    })

    it('should delete a config value', () => {
      cy.get('input[name="General.PublicURL"]')
        .clear()
        .should('be.empty')
      cy.get('button[data-cy="save"]').click()

      // save dialog
      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-General.PublicURL"] p[data-cy="old"]',
      )
        .should('contain', cfg.General.PublicURL)
        .siblings('p[data-cy="new"]')
        .should('not.exist')
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
        .should('be.empty')
        .type(newVal)
        .should('have.value', newVal)
      cy.get('button[data-cy="save"]').click()

      // save dialog
      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-General.PublicURL"] p[data-cy="old"]',
      )
        .should('contain', cfg.General.PublicURL)
        .siblings('p[data-cy="new"]')
        .should('contain', newVal)
      cy.get('button[type="button"]')
        .contains('Cancel')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close

      // value typed in
      cy.get('input[name="General.PublicURL"]').should('have.value', newVal)
      // reload page
      cy.reload()
      // since cancelled, data should be the same as before
      cy.get('input[name="General.PublicURL"]').should(
        'have.value',
        cfg.General.PublicURL,
      )
    })

    it('should reset pending config value changes', () => {
      const domain1 = c.domain()
      const domain2 = c.domain()

      cy.get('input[name="General.PublicURL"]')
        .clear()
        .should('be.empty')
        .type(domain1)
        .should('have.value', domain1)
      cy.get('input[name="Mailgun.EmailDomain"]')
        .type(domain2)
        .should('have.value', domain2)

      cy.get('button[data-cy="reset"]').click()

      cy.get('input[name="General.PublicURL"]').should(
        'have.value',
        cfg.General.PublicURL,
      )
      cy.get('input[name="Mailgun.EmailDomain"]').should('have.value', '')
    })

    it('should update a boolean toggle field', () => {
      cy.get('input[name="Twilio.Enable"]').check()
      cy.get('button[data-cy="save"]').click()

      // save dialog
      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-Twilio.Enable"] p[data-cy="old"]',
      )
        .should('contain', 'false')
        .siblings('p[data-cy="new"]')
        .should('contain', 'true')
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

      cy.get('input[name="Auth.RefererURLs-new-item"]')
        .clear()
        .should('be.empty')
        .type(domain1)
        .should('have.value', domain1)
      cy.get('input[name="Auth.RefererURLs-new-item"]')
        .clear()
        .should('be.empty')
        .type(domain2)
        .should('have.value', domain2)
      cy.get('input[name="Auth.RefererURLs-new-item"]')
        .clear()
        .should('be.empty')
        .type(domain3)
        .should('have.value', domain3)
      cy.get('button[data-cy="save"]').click()

      // save dialog
      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-Auth.RefererURLs"] p[data-cy="old"]',
      ).should('not.exist')
      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-Auth.RefererURLs"] p[data-cy="new"]',
      ).should('contain', `${domain1}, ${domain2}, ${domain3}`)
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
        .should('be.empty')
        .type(newKey)
        .should('have.value', newKey)
      cy.get('button[data-cy="save"]').click()

      // save dialog
      cy.get(
        'ul[data-cy="confirmation-diff"] li[data-cy="diff-Mailgun.APIKey"] p[data-cy="old"]',
      )
        .should('contain', cfg.Mailgun.APIKey)
        .siblings('p[data-cy="new"]')
        .should('contain', newKey)
      cy.get('button[type="submit"]')
        .contains('Confirm')
        .click()
      cy.get('div[role="document"]').should('not.exist') // dialog should close

      cy.get('input[name="Mailgun.APIKey"]').should('have.value', newKey)
    })
  })
}
