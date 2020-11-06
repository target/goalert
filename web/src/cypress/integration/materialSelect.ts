import { Chance } from 'chance'
import { testScreen } from '../support'
const c = new Chance()

function testMaterialSelect(): void {
  describe('Clear Fields', () => {
    describe('Escalation Policy Steps', () => {
      let ep: EP
      beforeEach(() => {
        cy.createEP().then((e: EP) => {
          ep = e
          return cy.visit(`escalation-policies/${ep.id}`)
        })
      })
      it('should clear fields and not reset with last values', () => {
        cy.fixture('users').then((users) => {
          const u1 = users[0]
          const u2 = users[1]

          cy.pageFab()
          cy.dialogTitle('Create Step')

          cy.get('button[data-cy="users-step"]').click()
          cy.dialogForm({ users: [u1.name, u2.name] })

          // Should clear field
          cy.dialogForm({ users: '' })
          cy.get(`[role=dialog] #dialog-form input[name="rotations"]`)
            .should('not.contain', u1.name)
            .should('not.contain', u2.name)

          cy.get(
            `[role=dialog] #dialog-form input[name="delayMinutes"]`,
          ).click()

          // Field should remian clear
          cy.get(`[role=dialog] #dialog-form input[name="rotations"]`)
            .should('not.contain', u1.name)
            .should('not.contain', u2.name)

          cy.dialogFinish('Submit')
        })
      })
    })
  })
  describe('Clear Required Fields', () => {
    describe('Escalation Policy', () => {
      it('should clear EP repeat count, reset with default value', () => {
        const defaultVal = '3'
        cy.visit('escalation-policies')
        cy.pageFab()
        cy.dialogTitle('Create Escalation Policy')

        // Clears field
        cy.dialogForm({ repeat: '' })
        cy.get('[role=dialog] #dialog-form input[name="repeat"]')
          .should('not.have.value', defaultVal)
          .blur()

        // Default value returns
        cy.get('[role=dialog] #dialog-form').click()
        cy.get('[role=dialog] #dialog-form input[name="repeat"]').should(
          'have.value',
          defaultVal,
        )

        cy.dialogFinish('Cancel')
      })
      it('should clear EP repeat count, reset with last value', () => {
        const name = 'SM EP ' + c.word({ length: 7 })
        const description = c.word({ length: 9 })
        const repeat = c.integer({ min: 0, max: 5 }).toString()

        cy.visit('escalation-policies')
        cy.pageFab()
        cy.dialogTitle('Create Escalation Policy')
        cy.dialogForm({ name, description, repeat })

        // Clears field
        cy.dialogForm({ repeat: '' })
        cy.get('[role=dialog] #dialog-form input[name="repeat"]').should(
          'not.have.value',
          repeat,
        )

        // Last value returns
        cy.get('[role=dialog] #dialog-form').click()
        cy.get('[role=dialog] #dialog-form input[name="repeat"]').should(
          'have.value',
          repeat,
        )

        // Should be on details page
        cy.dialogFinish('Submit')
        cy.get('body').should('contain', name).should('contain', description)
      })
    })
  })

  describe('render options', () => {
    it('should show selected value as an option in single-select mode', () => {
      cy.visit('/wizard')

      cy.form({ 'primarySchedule.timeZone': 'America/Chicago' })

      cy.get(`input[name="primarySchedule.timeZone"]`)
        .should('have.value', 'America/Chicago')
        .click()

      cy.get('[data-cy=select-dropdown]')
        .should('exist')
        .contains('America/Chicago')
    })
  })
}

testScreen('Material Select', testMaterialSelect)
