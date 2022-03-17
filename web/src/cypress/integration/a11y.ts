import { testScreen } from '../support'

function testA11y(): void {
  describe('no detectable a11y violations', () => {
    it('alerts list', () => {
      cy.visit('/alerts')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('rotations list', () => {
      cy.visit('/rotations')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('schedules list', () => {
      cy.visit('/schedules')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('escalation policies list', () => {
      cy.visit('/escalation-policies')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('services list', () => {
      cy.visit('/services')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('users list', () => {
      cy.visit('/users')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('profile', () => {
      cy.visit('/admin/profile')
      cy.injectAxe()
      cy.checkA11y()
    })

    // todo: failing, fix a11y issue
    it.skip('wizard', () => {
      cy.visit('/wizard')
      cy.injectAxe()
      cy.checkA11y()
    })

    // todo: failing, fix a11y issue
    it.skip('admin config', () => {
      cy.visit('/admin/config')
      cy.injectAxe()
      cy.checkA11y()
    })

    // todo: failing, fix a11y issue
    it.skip('admin system limits', () => {
      cy.visit('/admin/limits')
      cy.injectAxe()
      cy.checkA11y()
    })

    // todo: failing, fix a11y issue
    it.skip('admin toolbox', () => {
      cy.visit('/admin/toolbox')
      cy.injectAxe()
      cy.checkA11y()
    })

    it('admin message logs', () => {
      cy.visit('/admin/message-logs')
      cy.injectAxe()
      cy.checkA11y()
    })

    // todo: failing, fix a11y issue
    it.skip('api docs', () => {
      cy.visit('/docs')
      cy.injectAxe()
      cy.checkA11y()
    })
  })
}

testScreen('a11y', testA11y, false, true)
