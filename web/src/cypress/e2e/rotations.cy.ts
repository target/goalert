import { Chance } from 'chance'
import { testScreen } from '../support/e2e'
const c = new Chance()

function testRotations(screen: ScreenFormat): void {
  describe('List Page', () => {
    let rot: Rotation
    beforeEach(() => {
      cy.createRotation()
        .then((r: Rotation) => {
          rot = r
        })
        .visit('/rotations')
    })

    it('should handle searching', () => {
      // by name
      cy.pageSearch(rot.name)
      cy.get('body')
        .should('contain', rot.name)
        .should('contain', rot.description)
    })

    it('should link to details page', () => {
      cy.pageSearch(rot.name)
      cy.get('#app').contains(rot.name).click()
      cy.url().should('eq', Cypress.config().baseUrl + `/rotations/${rot.id}`)
    })

    describe('Creation', () => {
      it('should allow canceling', () => {
        if (screen === 'mobile') {
          cy.pageFab()
        } else {
          cy.get('button').contains('Create Rotation').click()
        }
        cy.dialogTitle('Create Rotation')
        cy.dialogFinish('Cancel')
      })
      ;['Hourly', 'Daily', 'Weekly', 'Monthly'].forEach((type) => {
        it(`should create a ${type} rotation when submitted`, () => {
          const name = 'SM Rot ' + c.word({ length: 8 })
          const description = c.word({ length: 10 })
          const tz = c.pickone(['America/Chicago', 'Africa/Accra', 'Etc/UTC'])
          const shiftLength = c.integer({ min: 1, max: 10 })

          if (screen === 'mobile') {
            cy.pageFab()
          } else {
            cy.get('button').contains('Create Rotation').click()
          }
          cy.dialogTitle('Create Rotation')
          cy.dialogForm({
            name,
            description,
            timeZone: tz,
            type,
            shiftLength: shiftLength.toString(),
            start: '2020-05-25T15:04',
          })
          cy.dialogFinish('Submit')

          // should be on details page
          cy.get('body')
            .should('contain', name)
            .should('contain', description)
            .should('contain', tz)
        })
      })

      describe('Hint', () => {
        it('should show handoff start time hint on certain dates', () => {
          const name = 'SM Rot ' + c.word({ length: 8 })
          const description = c.word({ length: 10 })
          const tz = c.pickone(['America/Chicago', 'Africa/Accra', 'Etc/UTC'])
          const shiftLength = c.integer({ min: 1, max: 10 })

          if (screen === 'mobile') {
            cy.pageFab()
          } else {
            cy.get('button').contains('Create Rotation').click()
          }
          cy.dialogTitle('Create Rotation')
          cy.dialogForm({
            name,
            description,
            timeZone: tz,
            type: 'Monthly',
            shiftLength: shiftLength.toString(),
            start: '2020-05-30T15:04',
          })
          cy.get('.MuiFormHelperText-root > .MuiTypography-root').should(
            'contain',
            'Unintended handoff behavior may occur when date starts after the 28th',
          )
        })
      })
    })
  })

  describe('Details Page', () => {
    let rot: Rotation
    beforeEach(() =>
      cy.createRotation({ numUsers: 3 }).then((r: Rotation) => {
        rot = r
        return cy.visit(`/rotations/${r.id}`)
      }),
    )

    it('should display users correctly', () => {
      cy.get('ul[data-cy=users]').find('li').as('parts')
      cy.get('@parts').eq(1).should('contain', rot.users[0].name)
      cy.get('@parts').eq(2).should('contain', rot.users[1].name)
    })

    it('should allow removing a user', () => {
      cy.get('ul[data-cy=users]').find('li').as('parts')

      // remove second user
      cy.get('@parts').eq(2).find('button').menu('Remove')

      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.get('@parts').should('not.contain', rot.users[1].name)

      // add again
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Add User').click()
      }
      cy.dialogTitle('Add User')
      cy.dialogForm({ users: rot.users[1].name })
      cy.dialogFinish('Submit')

      cy.get('ul[data-cy=users]')
        .find('li')
        .should('contain', rot.users[1].name)
    })

    it('should display users with the same name correctly when selecting users to add', () => {
      const name = 'John Smith'
      const email = 'johnSmith@test.com'
      const dupEmail = 'johnSmith2@test.com'
      cy.createUser({ name, email })
      cy.createUser({ name, email: dupEmail })

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Add User').click()
      }
      cy.dialogTitle('Add User')
      cy.get('input').click()
      cy.focused().type(name)

      cy.get('body').should('contain', email)
      cy.get('body').should('contain', dupEmail)

      cy.get('p').contains(email).click()
      cy.dialogFinish('Submit')
      cy.get('ul[data-cy=users]').find('li').should('contain', name)
    })

    it('should allow re-ordering participants', () => {
      // ensure list has fully loaded before drag/drop
      cy.get('ul[data-cy=users]').find('li').should('have.length', 4)

      // pick up a participant
      cy.get('[id="drag-0"]').focus()
      cy.focused().type('{enter}')
      cy.get('body').should(
        'contain',
        'Picked up sortable item 1. Sortable item 1 is in position 1 of 3',
      )

      // re-order
      cy.focused().type('{downarrow}', { force: true })
      cy.get('body').should(
        'contain',
        'Sortable item 1 was moved into position 2 of 3',
      )

      // place user, calls mutation
      cy.focused().type('{enter}', { force: true })
      cy.get('body').should(
        'contain',
        'Sortable item 1 was dropped at position 2 of 3',
      )

      cy.get('ul[data-cy=users]').find('li').as('parts')
      cy.get('@parts')
        .eq(1)
        .should('contain', rot.users[1].name)
        .should('not.contain', 'Shift ends')
      cy.get('@parts')
        .eq(2)
        .should('contain', rot.users[0].name)
        .should('contain', 'Shift ends')
      cy.get('@parts')
        .eq(3)
        .should('contain', rot.users[2].name)
        .should('not.contain', 'Shift ends')
    })

    it('should allow changing the active user', () => {
      cy.get('ul[data-cy=users]').find('li').as('parts')

      cy.get('@parts').eq(2).find('button').menu('Set Active')

      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.get('@parts')
        .eq(1)
        .should('contain', rot.users[0].name)
        .should('not.contain', 'Shift ends')
      cy.get('@parts')
        .eq(2)
        .should('contain', rot.users[1].name)
        .should('contain', 'Shift ends')
      cy.get('@parts')
        .eq(3)
        .should('contain', rot.users[2].name)
        .should('not.contain', 'Shift ends')
    })

    it('should allow deleting the rotation', () => {
      cy.get('[data-cy="card-actions"]')
        .find('button[aria-label="Delete"]')
        .click()

      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.url().should('eq', Cypress.config().baseUrl + '/rotations')
      cy.pageSearch(rot.name)
      cy.get('body').should('contain', 'No results')
    })

    it('should allow editing a rotation', () => {
      cy.createRotation({ shiftLength: 3, type: 'daily' }).then(
        (r: Rotation) => {
          const newName = c.word({ length: 15 })
          const newDesc = c.sentence({ words: 3 })
          const newTz = 'Africa/Accra'
          const invalidName = 'a'

          cy.visit(`/rotations/${r.id}`)
          cy.get('[data-cy="card-actions"]')
            .find('button[aria-label="Edit"]')
            .click()

          cy.dialogTitle('Edit Rotation')

          cy.dialogForm({
            name: invalidName,
          })
          cy.dialogClick('Submit')
          cy.get('body').should('contain', 'Must be at least 2 characters')

          cy.dialogForm({
            name: newName,
            description: newDesc,
            timeZone: newTz,
            type: 'Weekly',
            shiftLength: '5',
          })
          cy.dialogFinish('Retry')

          cy.get('body')
            .should('contain', newName)
            .should('contain', newDesc)
            .should('contain', newTz)
        },
      )
    })
  })
}

testScreen('Rotations', testRotations)
