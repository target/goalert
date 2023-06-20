import { Chance } from 'chance'
import { testScreen } from '../support/e2e'
import users from '../fixtures/users.json'
const c = new Chance()

function testMaterialSelect(screen: ScreenFormat): void {
  it('should display options with punctuation', () => {
    cy.createRotation().then((r) => {
      const u = users[3]
      cy.visit(`rotations/${r.id}`)
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Add User').click()
      }
      cy.dialogTitle('Add User')
      cy.get('input[name=users]').click()
      cy.focused().type(u.name.replace('.', ' '))
      cy.get('div[role=presentation]').contains(u.name)
    })
  })

  describe('Clear Optional Fields', () => {
    describe('Escalation Policy Steps', () => {
      let ep: EP
      beforeEach(() => {
        cy.createEP().then((e: EP) => {
          ep = e
          return cy.visit(`escalation-policies/${ep.id}`)
        })
      })

      it('should clear fields and not reset with last values', () => {
        const u1 = users[0]
        const u2 = users[1]

        cy.pageFab()
        cy.dialogTitle('Create Step')

        // populate users
        cy.get('button[data-cy="users-step"]').click()
        cy.dialogForm({ users: [u1.name, u2.name] })

        // clear field
        cy.dialogForm({ users: '' })
        cy.get(`input[name="users"]`)
          .should('not.contain', u1.name)
          .should('not.contain', u2.name)

        // unfocus
        cy.get(`input[name="users"]`).blur()
        cy.get(`input[name="users"]`)
          .should('not.contain', u1.name)
          .should('not.contain', u2.name)

        cy.dialogFinish('Submit')
      })
    })
  })

  describe('Clear Required Fields', () => {
    describe('Escalation Policy', () => {
      it('should clear EP repeat count, reset with default value', () => {
        const defaultVal = '3'
        cy.visit('escalation-policies')

        if (screen === 'mobile') {
          cy.pageFab()
        } else {
          cy.get('button').contains('Create Escalation Policy').click()
        }

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

        if (screen === 'mobile') {
          cy.pageFab()
        } else {
          cy.get('button').contains('Create Escalation Policy').click()
        }

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
}

testScreen('Material Select', testMaterialSelect)
