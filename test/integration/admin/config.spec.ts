import { test, expect } from '@playwright/test'
import { userSessionFile } from '../lib'
import Chance from 'chance'
const c = new Chance()

test('update config values', async ({ page, browser }) => {
  // const newURL = 'http://' + c.domain()
  // const newDomain = c.domain()
  // const newAPIKey = 'key-' + c.string({ length: 32, pool: '0123456789abcdef' })
  // cy.form({
  //   'General.PublicURL': newURL,
  //   'Mailgun.EmailDomain': newDomain,
  //   'Twilio.FromNumber': '',
  //   'Mailgun.APIKey': newAPIKey,
  //   'Twilio.Enable': false,
  // })
  // cy.get('button[data-cy=save]').click()
  // cy.dialogTitle('Apply Configuration Changes?')
  // cy.dialogFinish('Cancel')
  // cy.get('button[data-cy=save]').click()
  // cy.dialogTitle('Apply Configuration Changes?')
  // cy.dialogContains('-' + cfg.General.PublicURL)
  // cy.dialogContains('-' + cfg.Twilio.FromNumber)
  // cy.dialogContains('-' + cfg.Mailgun.APIKey)
  // cy.dialogContains('-true')
  // cy.dialogContains('+false')
  // cy.dialogContains('+' + newURL)
  // cy.dialogContains('+' + newDomain)
  // cy.dialogContains('+' + newAPIKey)
  // cy.dialogFinish('Confirm')
  // cy.get('input[name="General.PublicURL"]').should('have.value', newURL)
  // cy.get('input[name="Mailgun.EmailDomain"]').should('have.value', newDomain)
  // cy.get('input[name="Twilio.FromNumber"]').should('have.value', '')
  // cy.get('input[name="Mailgun.APIKey"]').should('have.value', newAPIKey)
  // cy.get('input[name="Twilio.Enable"]').should('not.be.checked')
})

test('reset pending config value changes', async ({ page, browser }) => {
  // const domain1 = c.domain()
  // const domain2 = c.domain()
  // cy.form({
  //   'General.PublicURL': domain1,
  //   'Mailgun.EmailDomain': domain2,
  // })
  // cy.get('button[data-cy="reset"]').click()
  // cy.get('input[name="General.PublicURL"]').should(
  //   'have.value',
  //   cfg.General.PublicURL,
  // )
  // cy.get('input[name="Mailgun.EmailDomain"]').should('have.value', '')
})

test('update a string list field', async ({ page, browser }) => {
  // const domain1 = 'http://' + c.domain()
  // const domain2 = 'http://' + c.domain()
  // const domain3 = cfg.General.PublicURL
  // cy.form({ 'Auth.RefererURLs-new-item': domain1 })
  // cy.form({ 'Auth.RefererURLs-new-item': domain2 })
  // cy.form({ 'Auth.RefererURLs-new-item': domain3 })
  // cy.get('button[data-cy=save]').click()
  // cy.dialogFinish('Confirm')
  // cy.get('input[name="Auth.RefererURLs-0"]').should('have.value', domain1)
  // cy.get('input[name="Auth.RefererURLs-1"]').should('have.value', domain2)
  // cy.get('input[name="Auth.RefererURLs-2"]').should('have.value', domain3)
  // cy.get('input[name="Auth.RefererURLs-new-item"]').should('have.value', '')
})

test('generate a slack manifest', async ({ page, browser }) => {
  // const publicURL = cfg.General.PublicURL
  // cy.get('[id="accordion-Slack"]').click()
  // cy.get('button').contains('App Manifest').click()
  // cy.dialogTitle('App Manifest')
  // // verify data integrity from pulled values
  // cy.dialogContains("name: 'GoAlert'")
  // cy.dialogContains(`request_url: '${publicURL}/api/v2/slack/message-action'`)
  // cy.dialogContains(
  //   `message_menu_options_url: '${publicURL}/api/v2/slack/menu-options'`,
  // )
  // cy.dialogContains(`'${publicURL}/api/v2/identity/providers/oidc/callback'`)
  // cy.dialogContains("display_name: 'GoAlert'")
  // // verify button routing to slack config page
  // cy.get('[data-cy="configure-in-slack"]')
  //   .should('have.attr', 'href')
  //   .and('contain', 'https://api.slack.com/apps?new_app=1&manifest_yaml=')
})
