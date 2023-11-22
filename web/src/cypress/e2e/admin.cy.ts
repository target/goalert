import { Chance } from 'chance'
import { DateTime } from 'luxon'
import { DebugMessage } from '../../schema'
import { testScreen, login, Config, pathPrefix } from '../support/e2e'
const c = new Chance()

function testAdmin(): void {
  describe('Admin System Limits Page', () => {
    let limits: Limits = new Map()
    beforeEach(() => {
      cy.getLimits().then((l: Limits) => {
        limits = l
        return cy.visit('/admin/limits')
      })
    })

    it('should allow updating system limits values', () => {
      cy.get('nav li').contains('System Limits').should('be.visible')

      const newContactMethods = c.integer({ min: 15, max: 1000 }).toString()
      const newEPActions = c.integer({ min: 15, max: 1000 }).toString()

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
        .then((curCfg: Config) => {
          cfg = curCfg
          return cy.visit('/admin').get('button[data-cy=save]').should('exist')
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

    it('should generate a slack manifest', () => {
      const publicURL = cfg.General.PublicURL
      cy.get('[id="accordion-Slack"]').click()
      cy.get('button').contains('App Manifest').click()
      cy.dialogTitle('App Manifest')

      // verify data integrity from pulled values
      cy.dialogContains("name: 'GoAlert'")
      cy.dialogContains(
        `request_url: '${publicURL}/api/v2/slack/message-action'`,
      )
      cy.dialogContains(
        `message_menu_options_url: '${publicURL}/api/v2/slack/menu-options'`,
      )
      cy.dialogContains(
        `'${publicURL}/api/v2/identity/providers/oidc/callback'`,
      )
      cy.dialogContains("display_name: 'GoAlert'")

      // verify button routing to slack config page
      cy.get('[data-cy="configure-in-slack"]')
        .should('have.attr', 'href')
        .and('contain', 'https://api.slack.com/apps?new_app=1&manifest_yaml=')
    })
  })

  describe('Admin Alert Count Page', () => {
    let svc1: Service
    let svc2: Service

    beforeEach(() => {
      cy.setTimeSpeed(0)
      cy.fastForward('-21h')

      cy.createService().then((s1: Service) => {
        svc1 = s1
        cy.createAlert({ serviceID: s1.id })
      })
      cy.createService().then((s2: Service) => {
        svc2 = s2
        cy.createAlert({ serviceID: s2.id })
        cy.createAlert({ serviceID: s2.id })
      })

      cy.fastForward('21h')
      cy.setTimeSpeed(1) // resume the flow of time

      return cy.visit('/admin/alert-counts')
    })

    it('should display alert counts', () => {
      const now = DateTime.local()
        .minus({ hours: 21 })
        .set({ minute: 0 })
        .toLocaleString({
          year: 'numeric',
          month: 'short',
          day: 'numeric',
          hour: 'numeric',
          minute: 'numeric',
        })

      cy.get(`.recharts-line-dots circle[r=3]`).last().trigger('mouseover')
      cy.get('[data-cy=alert-count-graph]')
        .should('contain', now)
        .should('contain', `${svc1.name}: 1`)
        .should('contain', `${svc2.name}: 2`)

      cy.get('[data-cy=alert-count-table]')
        .should('contain', svc1.name)
        .should('contain', svc2.name)
    })
  })

  describe('Admin Message Logs Page', () => {
    let debugMessage: DebugMessage

    before(() => {
      login() // required for before hooks

      cy.createOutgoingMessage({
        createdAt: DateTime.local().toISO(),
        serviceName: 'Test Service',
        userName: 'Test User',
      }).then((msg: DebugMessage) => {
        debugMessage = msg
      })
    })
    beforeEach(() => {
      cy.visit('/admin/message-logs')
    })

    it('should view the logs list and graph with one log', () => {
      const now = DateTime.local().toLocaleString({
        month: 'short',
        day: 'numeric',
      })

      cy.get(`.recharts-line-dots circle[value=1]`).trigger('click')
      cy.get('[data-cy=message-log-tooltip]')
        .should('contain', now)
        .should('contain', 'Count: 1')

      cy.get('[data-cy="paginated-list"]').as('list')
      cy.get('@list').should('have.length', 1)
      cy.get('@list')
        .eq(0)
        .should(
          'contain.text',
          DateTime.fromISO(debugMessage.createdAt).toFormat('fff'),
        )

      cy.get('@list')
        .eq(0)
        .should('contain.text', debugMessage.type + ' Notification')

      // todo: destination not supported: phone number value is pre-formatted
      // likely need to create a support function cy.getPhoneNumberInfo thru
      // gql to verify this info.

      cy.get('@list').eq(0).should('contain.text', debugMessage.serviceName)
      cy.get('@list').eq(0).should('contain.text', debugMessage.userName)
      cy.get('@list').eq(0).should('include.text', debugMessage.status) // "Failed" or "Failed (Permanent)" can exist
    })

    it('should segment the graph by service', () => {
      const now = DateTime.local().toLocaleString({
        month: 'short',
        day: 'numeric',
      })

      cy.get('[data-cy="spinner-loading"]').should('not.exist')
      cy.get('input[value="service"]').click()
      cy.get('[data-cy="spinner-loading"]').should('not.exist')

      cy.get('span[class="recharts-legend-item-text"]').should(
        'contain.text',
        'Test Service',
      )
      cy.get(`.recharts-line-dots circle[value=1]`).trigger('click')
      cy.get('[data-cy=message-log-tooltip]')
        .should('contain', now)
        .should('contain', 'Test Service: 1')
    })

    it('should segment the graph by user', () => {
      const now = DateTime.local().toLocaleString({
        month: 'short',
        day: 'numeric',
      })

      cy.get('[data-cy="spinner-loading"]').should('not.exist')
      cy.get('input[value="user"]').click()
      cy.get('[data-cy="spinner-loading"]').should('not.exist')

      cy.get('span[class="recharts-legend-item-text"]').should(
        'contain.text',
        'Test User',
      )
      cy.get(`.recharts-line-dots circle[value=1]`).trigger('click', {
        force: true,
      })
      cy.get('[data-cy=message-log-tooltip]')
        .should('contain', now)
        .should('contain', 'Test User: 1')
    })

    it('should select and view a logs details', () => {
      cy.get('[data-cy="paginated-list"]').eq(0).click()
      cy.get('[data-cy="debug-message-details"').as('details').should('exist')

      // todo: not asserting updatedAt, destination, or providerID
      cy.get('@details').should('contain.text', 'ID')
      cy.get('@details').should('contain.text', debugMessage.id)

      cy.get('@details').should('contain.text', 'Created At')
      cy.get('@details').should(
        'contain.text',
        DateTime.fromISO(debugMessage.createdAt).toFormat('fff'),
      )

      cy.get('@details').should('contain.text', 'Notification Type')
      cy.get('@details').should('contain.text', debugMessage.type)

      cy.get('@details').should('contain.text', 'Current Status')
      cy.get('@details').should('include.text', debugMessage.status)

      cy.get('@details').should('contain.text', 'User')
      cy.get('@details').should('contain.text', debugMessage.userName)

      cy.get('@details').should('contain.text', 'Service')
      cy.get('@details').should('contain.text', debugMessage.serviceName)

      cy.get('@details').should('contain.text', 'Alert')
      cy.get('@details').should('contain.text', debugMessage.alertID)
    })

    it('should verify user link from a logs details', () => {
      cy.get('[data-cy="paginated-list"]').eq(0).click()
      cy.get('[data-cy="debug-message-details"')
        .find('a')
        .contains(debugMessage?.userName ?? '')
        .should(
          'have.attr',
          'href',
          pathPrefix() + '/users/' + debugMessage.userID,
        )
        .should('have.attr', 'target', '_blank')
        .should('have.attr', 'rel', 'noopener noreferrer')
    })

    it('should verify service link from a logs details', () => {
      cy.get('[data-cy="paginated-list"]').eq(0).click()
      cy.get('[data-cy="debug-message-details"')
        .find('a')
        .contains(debugMessage?.serviceName ?? '')
        .should(
          'have.attr',
          'href',
          pathPrefix() + '/services/' + debugMessage.serviceID,
        )
        .should('have.attr', 'target', '_blank')
        .should('have.attr', 'rel', 'noopener noreferrer')
    })
  })
}

testScreen('Admin', testAdmin, false, true)
