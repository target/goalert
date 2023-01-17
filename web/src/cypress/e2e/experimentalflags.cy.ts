import { testScreen, testScreenWithFlags } from '../support/e2e'

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
