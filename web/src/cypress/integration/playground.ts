import { testScreen } from '../support'

function testPlayground(): void {
  describe('Creation', () => {
    beforeEach(() => {
      cy.visit('/api/graphql/explore')
    })

    it('should show docs', () => {
      // open docs
      cy.get('div').contains('Docs').click()
      cy.get('body').should('contain', 'user(...): User')

      // close docs
      cy.get('div').contains('Schema').click()
    })

    it('should show schema', () => {
      // open schema
      cy.get('div').contains('Schema').click()
      cy.get('body').should('contain', 'type Alert {')

      // close schema
      cy.get('div').contains('Schema').click()
    })
  })
}

testScreen('Playground', testPlayground)
