import { testScreen } from '../support/e2e'
import users from '../fixtures/users.json'

function testMaterialSelect(screen: ScreenFormat): void {
  let rot: Rotation
  beforeEach(() => {
    cy.createRotation().then((r: Rotation) => {
      rot = r
    })
  })

  it('should display options with punctuation', () => {
    const u = users[3]
    cy.visit(`rotations/${rot.id}`)

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

  it('should clear optional chips with multiple=true', () => {
    const u1 = users[0]
    const u2 = users[1]
    cy.visit(`rotations/${rot.id}`)

    if (screen === 'mobile') {
      cy.pageFab()
    } else {
      cy.get('button').contains('Add User').click()
    }

    cy.dialogTitle('Add User')

    // populate users
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
  })

  it('should clear required fields and reset with default value', () => {
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
}

testScreen('Material Select', testMaterialSelect)
