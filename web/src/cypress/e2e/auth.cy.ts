import { testScreen } from '../support/e2e'

import prof from '../fixtures/profile.json'

function testAuth(): void {
  before(() => {
    cy.clearCookies()
    cy.resetConfig()
    cy.visit('/')
  })

  it('should authenticate a user', () => {
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
}

testScreen('Auth', testAuth, true)
