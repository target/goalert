import { testScreen } from '../support'

function testSidebar(): void {
  beforeEach(() => cy.visit('/'))

  const testLink = (label: string, path: string): void => {
    it(`should have a link to ${label}`, () => {
      cy.pageNav(label)
      cy.url().should('eq', Cypress.config().baseUrl + path)
    })
  }

  testLink('Alerts', '/alerts')
  testLink('Rotations', '/rotations')
  testLink('Schedules', '/schedules')
  testLink('Escalation Policies', '/escalation-policies')
  testLink('Services', '/services')
  testLink('Users', '/users')
}

testScreen('Sidebar', testSidebar)
