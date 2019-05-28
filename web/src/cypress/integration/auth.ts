import { testScreen } from '../support'

testScreen('Auth', testAuth, true)

function testAuth(screen: ScreenFormat) {
  before(() =>
    cy
      .clearCookies()
      .resetConfig()
      .visit('/'),
  )

  it('should authenticate a user', () => {
    cy.fixture('profile').then(prof => {
      cy.get('form#auth-basic').as('form')

      cy.get('@form')
        .find('input[name=username]')
        .type(prof.username)
      cy.get('@form')
        .find('input[name=password]')
        .type(prof.password)
      cy.get('@form')
        .find('button[type=submit]')
        .click()

      cy.get('form#auth-basic').should('not.exist')
      cy.reload()
      cy.get('form#auth-basic').should('not.exist')

      cy.pageNav('Logout')

      cy.get('form#auth-basic').should('exist')
      cy.reload()
      cy.get('form#auth-basic').should('exist')
    })
  })
}
