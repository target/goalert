import { Chance } from 'chance'
import { testScreen, Limits, SystemLimits, Config } from '../support'
const c = new Chance()

testScreen('Admin', testAdmin, false, true)

function testAdmin() {
  describe('Admin System Limits Page', () => {
    let limits: Limits = new Map()
    beforeEach(() => {
      cy.getLimits().then(l => {
        limits = l
        return cy.visit('/admin/limits')
      })
    })

    it('should allow updating system limits values', () => {
      const newContactMethods = c.integer({ min: 0, max: 1000 }).toString()
      const newEPActions = c.integer({ min: 0, max: 1000 }).toString()

      const ContactMethodsPerUser = limits.get(
        'ContactMethodsPerUser',
      ) as SystemLimits
      const EPActionsPerStep = limits.get('EPActionsPerStep') as SystemLimits

      cy.form({
        ContactMethodsPerUser: newContactMethods,
        EPActionsPerStep: newEPActions,
      })

      cy.get('button[data-cy=save]').click()
      cy.dialogTitle('Apply Configuration Changes?')

      cy.dialogContains('-' + ContactMethodsPerUser.value)
      cy.dialogContains('-' + EPActionsPerStep.value)
      cy.dialogContains('+' + newContactMethods)
      cy.dialogContains('+' + newEPActions)
      cy.dialogFinish('Confirm')

      cy.get('input[name="EPActionsPerStep"]').should(
        'have.value',
        newEPActions,
      )
      cy.get('input[name="ContactMethodsPerUser"]').should(
        'have.value',
        newContactMethods,
      )
    })

    it('should reset pending system limit value changes', () => {
      const ContactMethodsPerUser = limits.get(
        'ContactMethodsPerUser',
      ) as SystemLimits
      const EPActionsPerStep = limits.get('EPActionsPerStep') as SystemLimits

      cy.form({
        ContactMethodsPerUser: c.integer({ min: 0, max: 1000 }).toString(),
        EPActionsPerStep: c.integer({ min: 0, max: 1000 }).toString(),
      })

      cy.get('button[data-cy="reset"]').click()

      cy.get('input[name="ContactMethodsPerUser"]').should(
        'have.value',
        ContactMethodsPerUser.value.toString(),
      )
      cy.get('input[name="EPActionsPerStep"]').should(
        'have.value',
        EPActionsPerStep.value.toString(),
      )
    })
  })
  describe('Admin Config Page', () => {
    let cfg: Config
    beforeEach(() => {
      return cy
        .resetConfig()
        .updateConfig({
          Mailgun: {
            APIKey: 'key-' + c.string({ length: 32, pool: '0123456789abcdef' }),
            EmailDomain: '',
          },
          Twilio: {
            Enable: true,
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

    it('should allow updating config values', () => {
      const newURL = 'http://' + c.domain()
      const newDomain = c.domain()
      const newAPIKey =
        'key-' + c.string({ length: 32, pool: '0123456789abcdef' })

      cy.form({
        'General.PublicURL': newURL,
        'Mailgun.EmailDomain': newDomain,
        'Twilio.FromNumber': '',
        'Mailgun.APIKey': newAPIKey,
        'Twilio.Enable': false,
      })
      cy.get('button[data-cy=save]').click()

      cy.dialogTitle('Apply Configuration Changes?')
      cy.dialogFinish('Cancel')

      cy.get('button[data-cy=save]').click()
      cy.dialogTitle('Apply Configuration Changes?')
      cy.dialogContains('-' + cfg.General.PublicURL)
      cy.dialogContains('-' + cfg.Twilio.FromNumber)
      cy.dialogContains('-' + cfg.Mailgun.APIKey)
      cy.dialogContains('-true')
      cy.dialogContains('+false')
      cy.dialogContains('+' + newURL)
      cy.dialogContains('+' + newDomain)
      cy.dialogContains('+' + newAPIKey)
      cy.dialogFinish('Confirm')

      cy.get('input[name="General.PublicURL"]').should('have.value', newURL)
      cy.get('input[name="Mailgun.EmailDomain"]').should(
        'have.value',
        newDomain,
      )
      cy.get('input[name="Twilio.FromNumber"]').should('have.value', '')
      cy.get('input[name="Mailgun.APIKey"]').should('have.value', newAPIKey)
      cy.get('input[name="Twilio.Enable"]').should('not.be.checked')
    })

    it('should reset pending config value changes', () => {
      const domain1 = c.domain()
      const domain2 = c.domain()

      cy.form({
        'General.PublicURL': domain1,
        'Mailgun.EmailDomain': domain2,
      })

      cy.get('button[data-cy="reset"]').click()

      cy.get('input[name="General.PublicURL"]').should(
        'have.value',
        cfg.General.PublicURL,
      )
      cy.get('input[name="Mailgun.EmailDomain"]').should('have.value', '')
    })

    it('should update a string list field', () => {
      const domain1 = 'http://' + c.domain()
      const domain2 = 'http://' + c.domain()
      const domain3 = cfg.General.PublicURL

      cy.form({ 'Auth.RefererURLs-new-item': domain1 })
      cy.form({ 'Auth.RefererURLs-new-item': domain2 })
      cy.form({ 'Auth.RefererURLs-new-item': domain3 })

      cy.get('button[data-cy=save]').click()
      cy.dialogFinish('Confirm')

      cy.get('input[name="Auth.RefererURLs-0"]').should('have.value', domain1)
      cy.get('input[name="Auth.RefererURLs-1"]').should('have.value', domain2)
      cy.get('input[name="Auth.RefererURLs-2"]').should('have.value', domain3)
      cy.get('input[name="Auth.RefererURLs-new-item"]').should('have.value', '')
    })
  })
}
