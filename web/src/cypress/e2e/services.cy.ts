import { Chance } from 'chance'
import { DateTime } from 'luxon'
import { testScreen } from '../support/e2e'
const c = new Chance()

function testServices(screen: ScreenFormat): void {
  beforeEach(() => {
    window.localStorage.setItem('show_services_new_feature_popup', 'false')
  })

  describe('Integration Keys', () => {
    let svc: Service
    beforeEach(() =>
      cy.createService().then((s: Service) => {
        svc = s
        return cy.visit(`/services/${svc.id}/integration-keys`)
      }),
    )

    const createKey = (type: string, name: string): void => {
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-key"]').click()
      }
      cy.dialogForm({ name, type })
      cy.dialogFinish('Submit')
    }

    it('should manage integration keys with mailgun disabled', () => {
      const domain = c.domain()
      const name = 'SM Int ' + c.word({ length: 8 })

      cy.get('body').should('contain', 'No integration keys')

      cy.updateConfig({
        Mailgun: {
          Enable: true,
          APIKey: 'key-' + c.string({ length: 32, pool: '0123456789abcdef' }),
          EmailDomain: domain,
        },
      })
      cy.reload()

      createKey('Email', name) // set email integration key
      cy.get('ul[data-cy=int-keys')
        .contains('li', name)
        .find('a')
        .should('have.attr', 'href')
        .and('include', 'mailto:')
        .and('include', domain)

      cy.updateConfig({ Mailgun: { Enable: false } })
      cy.reload()

      // check for disabled text
      cy.get('ul[data-cy=int-keys]').should(
        'contain',
        'Email integration keys are currently disabled.',
      )

      // check that dropdown type is hidden
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-key"]').click()
      }
      cy.get('input[name=type]')
        .findByLabel('Email')
        .should('have.attr', 'aria-disabled', 'true')
    })
  })

  describe('Labels', () => {
    beforeEach(() =>
      cy.createLabel().then((l: Label) => {
        return cy.visit(`/services/${l.svcID}/labels`)
      }),
    )

    it('should not be able to create a label when DisableLabelCreation is true', () => {
      const randomWord = c.word({
        length: 7,
      })
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-label"]').click()
      }
      cy.get('input[name=key]').findByLabel(`Create "${randomWord}"`)

      cy.updateConfig({
        General: {
          DisableLabelCreation: true,
        },
      })
      cy.reload()

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-label"]').click()
      }
      cy.get('input[name=key]').type(`Create "${randomWord}"`)
      cy.get('[data-cy="select-dropdown"]').should('contain', 'No options')
    })
  })

  describe('Metrics', () => {
    let closedAlert: Alert
    let openAlert: Alert
    beforeEach(() =>
      cy
        .setTimeSpeed(0)
        .fastForward('-25h')
        .createAlert()
        .then((a: Alert) => {
          closedAlert = a
          cy.fastForward('1m')
          cy.ackAlert(a.id)
          cy.fastForward('1m')
          cy.closeAlert(a.id)
          cy.fastForward('25h')
          cy.setTimeSpeed(1) // resume the flow of time
          // non-closed alert
          return cy.createAlert({ serviceID: a.serviceID })
        })
        .then((a: Alert) => {
          openAlert = a
          return cy.visit(`/services/${a.serviceID}/alert-metrics`)
        }),
    )

    it('should display alert metrics', () => {
      const now = DateTime.local().minus({ day: 1 }).toLocaleString({
        month: 'short',
        day: 'numeric',
      })

      // summary doesn't load by default on mobile (until scrolled to)
      cy.get('[data-cy=metrics-table]')
        .should('contain', closedAlert.id)
        .should('not.contain', openAlert.id)

      cy.get('path[name="Alert Count"]')
        .should('have.length', 1)
        .trigger('mouseover')
      cy.get('[data-cy=metrics-count-graph]')
        .should('contain', now)
        .should('contain', 'Alert Count: 1')
        .should('contain', 'Escalated: 0') // no ep steps

      cy.get(`.recharts-line-dots circle[r=3]`).last().trigger('mouseover')
      cy.get('[data-cy=metrics-averages-graph]')
        .should('contain', now)
        .should('contain', 'Avg. Ack: 1 min')
        .should('contain', 'Avg. Close: 2 min')
    })
  })
}

testScreen('Services', testServices)
