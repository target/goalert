import { testScreen } from '../support'

function testPlayground(): void {
  describe('Navigation', () => {
    beforeEach(() => {
      cy.visit('/api/graphql/explore')
    })

    it('should open, click around, and close docs', () => {
      cy.get('button[aria-label="Show Documentation Explorer"]').click()
      cy.get(
        '.graphiql-doc-explorer[aria-label="Documentation Explorer"]',
      ).should('be.visible')
      cy.get('.graphiql-doc-explorer').contains('Query').click()
      cy.get('.graphiql-doc-explorer').contains('a', 'alert').click()
      cy.get('button[aria-label="Hide Documentation Explorer"]').click()
      cy.get(
        '.graphiql-doc-explorer[aria-label="Documentation Explorer"]',
      ).should('not.exist')
    })
  })
}

testScreen('Playground', testPlayground)
