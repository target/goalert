/* eslint-disable camelcase */
declare namespace NodeJS {
  export interface Global {
    __webpack_public_path__: string
    pathPrefix: string
    applicationName: string
    GOALERT_VERSION: string
  }

  declare module '*.md'
  declare module '*.png'
}

declare module 'modernizr-esm/feature/inputtypes' {
  import * as m from 'modernizr'
  export const inputtypes = m.inputtypes
}
