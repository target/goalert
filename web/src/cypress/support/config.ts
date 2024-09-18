declare global {
  namespace Cypress {
    interface Chainable {
      getConfig: typeof getConfig

      /** Replaces the backend config entirely. */
      setConfig: typeof setConfig

      /** Merges new config values into existing backend config. */
      updateConfig: typeof updateConfig

      resetConfig: typeof resetConfig
    }
  }
}

interface General {
  PublicURL: string
  DisableLabelCreation: boolean
  NotificationDisclaimer: string
  DisableCalendarSubscriptions: boolean
  EnableV1GraphQL: boolean
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
    .then((res) => {
      expect(res.status, 'status code').to.eq(200)

      return JSON.parse(res.body) as Config
    })
}

function getConfig(): Cypress.Chainable<Config> {
  return cy.adminLogin(true).then((tok: string) => getConfigDirect(tok))
}

/*
 * setConfig replaces the current config completely with cfg
 */
function setConfig(cfg: ConfigInput): Cypress.Chainable<Config> {
  return cy.adminLogin(true).then((tok: string) =>
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
    (key) => (dst[key] = { ...(dst[key] || {}), ...src[key] }),
  )

  return dst
}

/*
 * updateConfig updates the config by merging the
 * provided newCfg values into the current config
 */
function updateConfig(newCfg: ConfigInput): Cypress.Chainable<Config> {
  return getConfig().then((cfg) => {
    return setConfig(merge(cfg, newCfg))
  })
}

function resetConfig(): Cypress.Chainable<Config> {
  const base = String(Cypress.config('baseUrl'))

  return setConfig({
    General: { PublicURL: base },
    Twilio: {
      Enable: true,
      AccountSID: 'fake-for-testing',
      AuthToken: 'fake-for-testing',
      FromNumber: '+16125550123',
    },
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
