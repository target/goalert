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

interface ConfigInput {
  [index: string]: any

  General?: {
    PublicURL?: string
    DisableLabelCreation?: boolean
    NotificationDisclaimer?: string
    DisableCalendarSubscriptions?: boolean
  }
  Auth?: {
    RefererURLs?: [string]
    DisableBasic?: boolean
  }
  Mailgun?: {
    Enable?: boolean
    APIKey?: string
    EmailDomain?: string
    DisableValidation?: boolean
  }
  Twilio?: {
    Enable?: boolean
    AccountSID?: string
    AuthToken?: string
    FromNumber?: string
  }
  Feedback?: {
    Enable?: boolean
    OverrideURL?: string
  }
}
export interface Config {
  [index: string]: any

  General: {
    PublicURL: string
    DisableLabelCreation: boolean
    NotificationDisclaimer: string
  }
  Auth: {
    RefererURLs: [string]
    DisableBasic: boolean
  }
  Mailgun: {
    Enable: boolean
    APIKey: string
    EmailDomain: string
  }
  Twilio: {
    Enable: boolean
    AccountSID: string
    AuthToken: string
    FromNumber: string
  }
  Feedback: {
    Enable: boolean
    OverrideURL: string
  }
  Slack?: {
    Enable?: boolean
    ClientID?: string
    ClientSecret?: string
    AccessToken?: string
  }
}

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

function merge(dst: any, src: ConfigInput): Config {
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
