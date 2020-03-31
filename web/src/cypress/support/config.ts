interface General {
  PublicURL: string
  DisableLabelCreation: boolean
  NotificationDisclaimer: string
  DisableCalendarSubscriptions: boolean
}

interface Auth {
  RefererURLs: [string]
  DisableBasic: boolean
}

interface Mailgun {
  Enable: boolean
  APIKey: string
  EmailDomain: string
}

interface Twilio {
  Enable: boolean
  AccountSID: string
  AuthToken: string
  FromNumber: string
}

interface Feedback {
  Enable: boolean
  OverrideURL: string
}

interface Slack {
  Enable: boolean
  ClientID: string
  ClientSecret: string
  AccessToken: string
}

export interface Config {
  [index: string]: Partial<General | Auth | Mailgun | Twilio | Feedback | Slack>
  General: General
  Auth: Auth
  Mailgun: Mailgun
  Twilio: Twilio
  Feedback: Feedback
  Slack: Slack
}

type ConfigInput = Pick<Config, string>

function getConfigDirect(token: string): Cypress.Chainable<Config> {
  return cy
    .request({
      url: '/api/v2/config',
      method: 'GET',
      auth: { bearer: token },
    })
    .then(res => {
      expect(res.status, 'status code').to.eq(200)

      return JSON.parse(res.body)
    })
}

function getConfig(): Cypress.Chainable<Config> {
  return cy.adminLogin(true).then(tok => getConfigDirect(tok))
}

function setConfig(cfg: ConfigInput): Cypress.Chainable<Config> {
  return cy.adminLogin(true).then(tok =>
    cy
      .request({
        url: '/api/v2/config',
        method: 'PUT',
        body: JSON.stringify(cfg),
        auth: { bearer: tok },
      })
      .then(() => getConfigDirect(tok)),
  )
}

function merge(dst: Config, src: ConfigInput): Config {
  Object.keys(src).forEach(
    key => (dst[key] = { ...(dst[key] || {}), ...src[key] }),
  )

  return dst
}

function updateConfig(newCfg: ConfigInput): Cypress.Chainable<Config> {
  return getConfig().then(cfg => {
    return setConfig(merge(cfg, newCfg))
  })
}

function resetConfig(): Cypress.Chainable<Config> {
  const base = String(Cypress.config('baseUrl'))

  return setConfig({
    General: { PublicURL: base },
    Slack: {
      Enable: true,
      ClientID: '000000000000.000000000000',
      ClientSecret: '00000000000000000000000000000000',
      AccessToken:
        'xoxp-000000000000-000000000000-000000000000-00000000000000000000000000000000',
    },
  })
}

Cypress.Commands.add('getConfig', getConfig)
Cypress.Commands.add('setConfig', setConfig)
Cypress.Commands.add('updateConfig', updateConfig)
Cypress.Commands.add('resetConfig', resetConfig)
