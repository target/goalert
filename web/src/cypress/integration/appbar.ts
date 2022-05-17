import { testScreen } from '../support'

function testAppbar(): void {
  it('should display correct breadcrumbs on list page', () => {
    cy.visit('/alerts')
    cy.get('[aria-label="Breadcrumbs"]').should('contain.text', 'alerts')

    cy.visit('/rotations')
    cy.get('[aria-label="Breadcrumbs"]').should('contain.text', 'rotations')

    cy.visit('/escalation-policies')
    cy.get('[aria-label="Breadcrumbs"]').should(
      'contain.text',
      'escalation policies',
    )

    cy.visit('/services')
    cy.get('[aria-label="Breadcrumbs"]').should('contain.text', 'services')

    cy.visit('/users')
    cy.get('[aria-label="Breadcrumbs"]').should('contain.text', 'users')
  })

  it('should display correct breadcrumbs on single pages', () => {
    cy.visit('/profile')
    cy.get('[aria-label="Breadcrumbs"]').should('contain.text', 'profile')

    cy.visit('/admin/config')
    cy.get('[aria-label="Breadcrumbs"]').should('include.text', 'Admin Details')

    cy.visit('/admin/limits')
    cy.get('[aria-label="Breadcrumbs"]').should('include.text', 'Admin Details')

    cy.visit('/admin/toolbox')
    cy.get('[aria-label="Breadcrumbs"]').should('include.text', 'Admin Details')

    cy.visit('/admin/message-logs')
    cy.get('[aria-label="Breadcrumbs"]').should('include.text', 'Admin Details')
  })

  it('should display correct breadcrumbs on details page', () => {})
  it('should display correct breadcrumbs on sub page', () => {})
}

testScreen('Appbar', testAppbar, false, true)
