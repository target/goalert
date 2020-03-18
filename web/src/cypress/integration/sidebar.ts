import { testScreen } from '../support'

testScreen('Sidebar', testSidebar)

function testSidebar() {
  beforeEach(() => cy.visit('/'))

  const testLink = (label: string, path: string) => {
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
  testLink('Profile', '/profile')

  describe('Feedback', () => {
    it('should not display by default', () => {
      cy.get('[data-cy=feedback-link]').should('not.exist')
    })
    it('should display with default href when enabled', () => {
      cy.updateConfig({ Feedback: { Enable: true } })
      cy.reload()
      cy.pageNav('Feedback', true)
      cy.get('[data-cy=feedback-link]')
        .should('have.attr', 'href')
        .and(
          'match',
          /https:\/\/www\.surveygizmo\.com\/s3\/4106900\/GoAlert-Feedback/,
        )
    })
    it('should display with correct href when overridden', () => {
      cy.updateConfig({
        Feedback: { Enable: true, OverrideURL: 'https://www.goalert.me' },
      })
      cy.reload()
      cy.pageNav('Feedback', true)
      cy.get('[data-cy=feedback-link]')
        .should('have.attr', 'href')
        .and('match', /https:\/\/www\.goalert\.me/)
    })
  })
}
