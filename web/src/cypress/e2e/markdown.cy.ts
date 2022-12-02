import { testScreen } from '../support'

// todo: test alert details markdown

function testMarkdownTables(): void {
  describe('Markdown Tables', () => {
    it('should render tables in html', () => {
      cy.visit('/docs')
      cy.get('table > thead > tr > th').should('exist').contains('Name')
    })
  })
}

testScreen('Markdown', testMarkdownTables)
