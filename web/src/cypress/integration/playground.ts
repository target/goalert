import { testScreen } from '../support'

function testPlayground(): void {
  describe('Navigation', () => {
    beforeEach(() => {
      cy.visit('/api/graphql/explore')
    })

    it('should open, click around, and close docs', () => {
      cy.get('.docExplorerShow').click()
      cy.get('.docExplorerShow').should('not.exist')
      cy.get('.doc-explorer').contains('Documentation Explorer')
      cy.get('.doc-explorer').contains('Query').click()
      cy.get('.doc-explorer').contains('alert').click()
      cy.get('.docExplorerHide').click()
      cy.get('.docExplorerShow').should('exist')
    })
  })
}

testScreen('Playground', testPlayground)
