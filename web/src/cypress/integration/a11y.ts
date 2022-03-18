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
      waitForPageLoad: () => void = () => {
        // eslint-disable-next-line cypress/no-unnecessary-waiting
        cy.wait(2000)
      },
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

    testRoute('alerts list', '/alerts?allServices=1&filter=all', () => {
      cy.get('ul[data-cy="apollo-list"]')
    })
    testRoute('rotations list', '/rotations', () => {
      cy.get('ul[data-cy="apollo-list"]')
    })
    testRoute('schedules list', '/schedules', () => {
      cy.get('ul[data-cy="apollo-list"]')
    })
    testRoute('escalation policies list', '/escalation-policies', () => {
      cy.get('ul[data-cy="apollo-list"]')
    })
    testRoute('services list', '/services', () => {
      cy.get('ul[data-cy="apollo-list"]')
    })
    testRoute('users list', '/users', () => {
      cy.get('ul[data-cy="apollo-list"]')
    })
    testRoute('profile', '/profile', () => {
      cy.get('ul[data-cy="contact-methods"]')
      cy.get('ul[data-cy="notification-rules"]')
      cy.get('div[data-cy="alert-status-updates"]')
    })
    testRoute('wizard', '/wizard')
    testRoute('admin config', '/admin/config')
    testRoute('admin system limits', '/admin/limits')
    testRoute('admin toolbox', '/admin/toolbox')
    testRoute('admin message logs', '/admin/message-logs')
    testRoute('api docs', '/docs')
  })
}

testScreen('a11y', testA11y, false, true)
