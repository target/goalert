declare global {
  namespace NodeJS {
    interface Global {
      pathPrefix: string
      GOALERT_VERSION: string
      Cypress?: any
    }
  }
}
declare var __webpack_public_path__: string
declare var __webpack_require__: any

export const pathPrefix = global.pathPrefix
// eslint-disable-next-line
__webpack_require__.p = __webpack_public_path__ = pathPrefix.replace(
  /\/?$/,
  '/',
)

export const GOALERT_VERSION = process.env.GOALERT_VERSION || 'dev'
global.GOALERT_VERSION = GOALERT_VERSION

export const isCypress = Boolean(global.Cypress)
