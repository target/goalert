// ***********************************************************
// This example support/index.js is processed and
// loaded automatically before your test files.
//
// This is a great place to put global configuration and
// behavior that modifies Cypress.
//
// You can change the location of this file or turn off
// automatically serving support files with the
// 'supportFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/configuration
// ***********************************************************

/// <reference types="cypress" />

import './slack'
import './admin'
import './alert'
import './calendar-subscription'
import './fail-fast'
import './service'
import './ep'
import './rotation'
import './graphql'
import './login'
import './profile'
import './schedule'
import './select-by-label'
import './menu'
import './navigate-to-and-from'
import './page-search'
import './page-action'
import './page-nav'
import './page-fab'
import './sql'
import './form'
import './dialog'
import './limits'

import { Settings } from 'luxon'
Settings.throwOnInvalid = true

declare module 'luxon' {
  interface TSSettings {
    throwOnInvalid: true
  }
}

Cypress.Keyboard.defaults({
  keystrokeDelay: 0,
})

export * from './config'
export * from './util'
