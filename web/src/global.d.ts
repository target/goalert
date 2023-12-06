/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable camelcase */
/* eslint-disable no-var */

declare namespace NodeJS {
  declare module '*.md'
  declare module '*.png'
  declare module '*.svg'
  declare module '*.gif'
}

var pathPrefix: string
var applicationName: string
var GOALERT_VERSION: string
var Cypress: any
