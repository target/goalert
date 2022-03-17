import { testScreen } from '../support'

function testA11y(): void {
  describe('A11y Compatibility', () => {
    beforeEach(() => {
      cy.visit('alerts/?poll=0')
      cy.injectAxe()
    })

    it('test page accessibility', () => {
      cy.checkA11y()
    })
  })
}

testScreen('A11y', testA11y, false, true)
