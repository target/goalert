import { Chance } from 'chance'
import { DebugMessage } from '../../schema'
import { testScreen, Config, pathPrefix } from '../support'
const c = new Chance()

function testAdmin(): void {
  describe('Admin System Limits Page', () => {
    let limits: Limits = new Map()
    beforeEach(() => {
      cy.getLimits().then((l: Limits) => {
        limits = l
        return cy.visit('/admin/limits')
      })
    })
  })

  describe('Admin Config Page', () => {
    let cfg: Config
    beforeEach(() => {
      return cy
        .resetConfig()
        .updateConfig({
          Mailgun: {
            APIKey: 'key-' + c.string({ length: 32, pool: '0123456789abcdef' }),
            EmailDomain: '',
          },
          Twilio: {
            Enable: true,
            AccountSID:
              'AC' + c.string({ length: 32, pool: '0123456789abcdef' }),
            AuthToken: c.string({ length: 32, pool: '0123456789abcdef' }),
            FromNumber: '+17633' + c.string({ length: 6, pool: '0123456789' }),
          },
        })
        .then((curCfg: Config) => {
          cfg = curCfg
          return cy.visit('/admin').get('button[data-cy=save]').should('exist')
        })
    })
  })

  describe('Admin Alert Count Page', () => {
    let svc1: Service
    let svc2: Service

    beforeEach(() => {
      cy.setTimeSpeed(0)
      cy.fastForward('-21h')

      cy.createService().then((s1: Service) => {
        svc1 = s1
        cy.createAlert({ serviceID: s1.id })
      })
      cy.createService().then((s2: Service) => {
        svc2 = s2
        cy.createAlert({ serviceID: s2.id })
        cy.createAlert({ serviceID: s2.id })
      })

      cy.fastForward('21h')
      cy.setTimeSpeed(1) // resume the flow of time

      return cy.visit('/admin/alert-counts')
    })
  })

  describe('Admin Message Logs Page', () => {
    let debugMessage: DebugMessage

    before(() => {
      cy.createOutgoingMessage().then((msg: DebugMessage) => {
        debugMessage = msg
        cy.visit('/admin/message-logs')
      })
    })
  })
}

testScreen('Admin', testAdmin, false, true)
