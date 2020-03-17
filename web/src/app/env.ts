declare global {
  namespace NodeJS {
    interface Global {
      pathPrefix: string
      GOALERT_VERSION: string
      Cypress?: any
    }
  }
}
declare let __webpack_public_path__: string
declare let __webpack_require__: any

export const pathPrefix = global.pathPrefix

if (typeof __webpack_require__ !== 'undefined')
  // eslint-disable-next-line
  __webpack_require__.p = __webpack_public_path__ = pathPrefix.replace(
    /\/?$/,
    '/',
  )

export const GOALERT_VERSION = process.env.GOALERT_VERSION || 'dev'
global.GOALERT_VERSION = GOALERT_VERSION

export const isCypress = Boolean(global.Cypress)
