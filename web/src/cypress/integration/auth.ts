import { testScreen } from '../support'

function testAuth(): void {
  before(() => cy.clearCookies().resetConfig().visit('/'))

  it('should authenticate a user', () => {
    cy.fixture('profile').then((prof) => {
      cy.form(
        {
          username: prof.username,
          password: prof.password,
        },
        'form#auth-basic',
      )
      cy.get('form#auth-basic').as('form')

      cy.get('button[type=submit]').click()

      cy.get('form#auth-basic').should('not.exist')
      cy.reload()
      cy.get('form#auth-basic').should('not.exist')

      cy.get('[aria-label="Manage Profile"]').click()
      cy.get('[data-cy="manage-profile"]')
        .find('button')
        .contains('Logout')
        .click()

      cy.get('form#auth-basic').should('exist')
      cy.reload()
      cy.get('form#auth-basic').should('exist')
    })
  })
}

testScreen('Auth', testAuth, true)
