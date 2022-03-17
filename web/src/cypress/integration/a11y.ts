import { testScreen } from '../support'

function testA11y(): void {
  describe('A11y Compatibility', () => {
    it('has no detectable a11y violations on alerts list', () => {
      cy.visit('/alerts')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('has no detectable a11y violations on rotations list', () => {
      cy.visit('/rotations')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('has no detectable a11y violations on schedules list', () => {
      cy.visit('/schedules')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('has no detectable a11y violations on escalation policies list', () => {
      cy.visit('/escalation-policies')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('has no detectable a11y violations on services list', () => {
      cy.visit('/services')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('has no detectable a11y violations on users list', () => {
      cy.visit('/users')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('has no detectable a11y violations on profile', () => {
      cy.visit('/admin/profile')
      cy.injectAxe()
      cy.checkA11y()
    })

    // todo: failing, fix a11y issue
    it.skip('has no detectable a11y violations on wizard', () => {
      cy.visit('/wizard')
      cy.injectAxe()
      cy.checkA11y()
    })

    // todo: failing, fix a11y issue
    it.skip('has no detectable a11y violations on admin config', () => {
      cy.visit('/admin/config')
      cy.injectAxe()
      cy.checkA11y()
    })

    // todo: failing, fix a11y issue
    it.skip('has no detectable a11y violations on admin system limits', () => {
      cy.visit('/admin/limits')
      cy.injectAxe()
      cy.checkA11y()
    })

    // todo: failing, fix a11y issue
    it.skip('has no detectable a11y violations on admin toolbox', () => {
      cy.visit('/admin/toolbox')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('has no detectable a11y violations on admin message logs', () => {
      cy.visit('/admin/message-logs')
      cy.injectAxe()
      cy.checkA11y()
    })

    // todo: failing, fix a11y issue
    it.skip('has no detectable a11y violations on api docs', () => {
      cy.visit('/docs')
      cy.injectAxe()
      cy.checkA11y()
    })
  })
}

testScreen('A11y', testA11y, false, true)
