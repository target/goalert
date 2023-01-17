import { testScreen, testScreenWithFlags } from '../support/e2e'

// These tests validate that the experimental flags are being set correctly
// during the Cypress tests. The testScreen and testScreenWithFlags functions
// are defined in web/src/cypress/support/util.ts

function testNoFlags(): void {
  beforeEach(() => cy.visit('/'))
  it('should return no flags', () => {
    cy.graphql('{experimentalFlags}').then((res) => {
      expect(res.experimentalFlags).to.deep.equal([])
    })
  })
}

function testExampleFlag(): void {
  beforeEach(() => cy.visit('/'))
  it('should return example flag', () => {
    cy.graphql('{experimentalFlags}').then((res) => {
      expect(res.experimentalFlags).to.deep.equal(['example'])
    })
  })
}

describe('Experimental Flags', () => {
  testScreen('Default', testNoFlags)
  testScreenWithFlags('Example Flag', testExampleFlag, ['example'])
})
