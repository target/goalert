import { testScreen } from '../support'

function testA11y(): void {
  describe('no detectable a11y violations', () => {
    before(() => {
      // create one of everything, all associated with each other so that every list will be populated
      cy.createService().then((service) => {
        cy.setScheduleTarget().then((scheduleTarget) => {
          cy.createEPStep({
            epID: service.epID,
            targets: [{ id: scheduleTarget.scheduleID, type: 'schedule' }],
          }).then(() => {
            cy.createManyAlerts(3, { serviceID: service.id })
          })
        })
      })
    })

    function testRoute(
      testName: string,
      route: string,
      waitForPageLoad: () => void,
    ): void {
      it(testName, () => {
        cy.visit(route)
        cy.injectAxe()
        waitForPageLoad()
        cy.checkA11y(undefined, {
          includedImpacts: ['critical'], // only report and assert for critical impact items
        })
      })
    }

    function waitForList(): Cypress.Chainable {
      return cy.get('ul[data-cy="apollo-list"]')
    }

    testRoute('alerts list', '/alerts?allServices=1&filter=all', waitForList)

    testRoute('profile', '/profile', () => {
      cy.get('ul[data-cy="contact-methods"]')
      cy.get('ul[data-cy="notification-rules"]')
      cy.get('div[data-cy="alert-status-updates"]')
    })

    testRoute('wizard', '/wizard', () => {
      cy.get('input[name="primarySchedule.timeZone"]')
    })

    testRoute('admin config', '/admin/config', () => {
      cy.get('div[data-cy="admin-config"]')
    })

    testRoute('admin system limits', '/admin/limits', () => {
      cy.get('div[data-cy="admin-limits"]')
    })

    testRoute('admin toolbox', '/admin/toolbox', () => {
      cy.get('div[data-cy="admin-toolbox"]')
    })

    testRoute('admin message logs', '/admin/message-logs', () => {
      cy.get('div[data-cy="admin-message-logs"]')
    })

    testRoute('api docs', '/docs', () => {
      cy.get('div[data-cy="api-docs"]')
    })
  })
}

testScreen('a11y', testA11y, false, true)
