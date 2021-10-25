/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable camelcase */
/* eslint-disable no-var */

declare namespace NodeJS {
  declare module '*.md'
  declare module '*.png'
}

var __webpack_public_path__: string
var pathPrefix: string
var applicationName: string
var GOALERT_VERSION: string
var Cypress: any

declare module 'modernizr-esm/feature/inputtypes' {
  import * as m from 'modernizr'
  export const inputtypes = m.inputtypes
}
