export const pathPrefix = global.pathPrefix || ''
export const applicationName = global.applicationName || 'GoAlert'

export const GOALERT_VERSION = global.GOALERT_VERSION || 'dev'

export const isCypress = Boolean(global.Cypress)

// read nonce from csp-nonce meta tag
export const nonce =
  document
    .querySelector('meta[property="csp-nonce"]')
    ?.getAttribute('content') || ''
