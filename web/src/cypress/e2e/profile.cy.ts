import { Chance } from 'chance'
import { testScreen } from '../support/e2e'
import profile from '../fixtures/profile.json'
const c = new Chance()

function testProfile(): void {
  it('should list and link on-call services', () => {
    const name = 'SVC ' + c.word({ length: 8 })

    return cy
      .createService({ name })
      .then((svc: Service) => {
        return cy
          .createEPStep({
            epID: svc.epID,
            targets: [{ type: 'user', id: profile.id }],
          })
          .engineTrigger()
          .then(() => svc.id)
      })
      .then((svcID: string) => {
        cy.get('body').contains('a', 'On-Call').click()

        cy.get('body').contains('a', name).click()

        cy.url().should('eq', Cypress.config().baseUrl + '/services/' + svcID)
      })
  })

  describe('Settings', () => {
    it('should visit profile', () => {
      cy.visit('/')
      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy="manage-profile"]')
        .find('button')
        .contains('Manage Profile')
        .click()
      cy.url().should('eq', Cypress.config().baseUrl + '/users/' + profile.id)
    })

    it('should change the theme mode and color', () => {
      cy.get('[aria-label="Manage Profile"]').click()

      // test changing theme color
      let appbarColor: string
      cy.get('[data-cy="manage-profile"] button').contains('Light').click()
      cy.get('[data-cy="app-bar"]').then(
        (el) => (appbarColor = el.css('background-color')),
      )

      // set input of color
      cy.get(
        '[data-cy="manage-profile"] button[aria-label="More Options"]',
      ).click()
      cy.get('input[id="custom-color-picker"]')
        .invoke('val', '#fff000')
        .trigger('input')

      // assert primary color has changed
      cy.reload()
      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy="app-bar"]').then((el) =>
        expect(appbarColor).not.to.equal(el.css('background-color')),
      )

      // test changing theme mode to dark
      cy.get('[data-cy="manage-profile"] button').contains('Dark').click()

      // assert theme mode has changed
      cy.reload()
      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy="app-bar"]').then((el) =>
        expect(appbarColor).not.to.equal(el.css('background-color')),
      )
    })

    it('should not display feedback by default', () => {
      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy=feedback]').should('not.exist')
    })

    it('should display feedback with default href when enabled', () => {
      cy.updateConfig({ Feedback: { Enable: true } })
      cy.reload()
      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy="manage-profile"]')
        .find('[data-cy=feedback]')
        .should('have.attr', 'href')
        .and(
          'match',
          /https:\/\/www\.surveygizmo\.com\/s3\/4106900\/GoAlert-Feedback/,
        )
    })

    it('should display feedback with correct href when overridden', () => {
      cy.updateConfig({
        Feedback: { Enable: true, OverrideURL: 'https://www.goalert.me' },
      }).then(() => {
        cy.get('[aria-label="Manage Profile"]').click()
        cy.get('[data-cy="manage-profile"]')
          .find('[data-cy=feedback]')
          .should('have.attr', 'href')
          .and('match', /https:\/\/www\.goalert\.me/)
      })
    })
  })
}

testScreen('Profile', testProfile)
