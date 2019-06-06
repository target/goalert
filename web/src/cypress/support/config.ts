declare namespace Cypress {
  interface Chainable {
    getConfig: typeof getConfig

    /** Replaces the backend config entirely. */
    setConfig: typeof setConfig

    /** Merges new config values into existing backend config. */
    updateConfig: typeof updateConfig

    resetConfig: typeof resetConfig
  }
}

interface ConfigInput {
  [index: string]: any

  General?: {
    PublicURL?: String
    DisableLabelCreation?: Boolean
    NotificationDisclaimer?: string
  }
  Auth?: {
    RefererURLs?: [String]
    DisableBasic?: Boolean
  }
  Mailgun?: {
    Enable?: Boolean
    APIKey?: String
    EmailDomain?: String
    DisableValidation?: Boolean
  }
  Twilio?: {
    Enable?: Boolean
    AccountSID?: String
    AuthToken?: String
    FromNumber?: String
  }
  Feedback?: {
    Enable?: Boolean
    OverrideURL?: String
  }
}
interface Config {
  [index: string]: any

  General: {
    PublicURL: string
    DisableLabelCreation: Boolean
    NotificationDisclaimer: string
  }
  Auth: {
    RefererURLs: [string]
    DisableBasic: Boolean
  }
  Mailgun: {
    Enable: Boolean
    APIKey: string
    EmailDomain: string
  }
  Twilio: {
    Enable: Boolean
    AccountSID: String
    AuthToken: String
    FromNumber: String
  }
  Feedback: {
    Enable: Boolean
    OverrideURL: String
  }
  Slack?: {
    Enable?: Boolean
    ClientID?: String
    ClientSecret?: String
    AccessToken?: String
  }
}

function getConfigDirect(token: String): Cypress.Chainable<Config> {
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
      ClientID: '555449060693.555449060694',
      ClientSecret: '52fdfc072182654f163f5f0f9a621d72',
      AccessToken:
        'xoxp-555449060693-555449060694-587071460694-9566c74d10037c4d7bbb0407d1e2c649',
    },
  })
}

Cypress.Commands.add('getConfig', getConfig)
Cypress.Commands.add('setConfig', setConfig)
Cypress.Commands.add('updateConfig', updateConfig)
Cypress.Commands.add('resetConfig', resetConfig)
