import { Chance } from 'chance'
import { DateTime } from 'luxon'

import { testScreen } from '../support'
const c = new Chance()

testScreen('Rotations', testRotations)

function testRotations() {
  describe('List Page', () => {
    let rot: Rotation
    beforeEach(() => {
      cy.createRotation()
        .then(r => {
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
      cy.get('#app')
        .contains(rot.name)
        .click()
      cy.url().should('eq', Cypress.config().baseUrl + `/rotations/${rot.id}`)
    })

    describe('Creation', () => {
      it('should allow canceling', () => {
        cy.pageFab()
        cy.dialogTitle('Create Rotation')
        cy.dialogFinish('Cancel')
      })
      ;['Hourly', 'Daily', 'Weekly'].forEach(type => {
        it(`should create a ${type} rotation when submitted`, () => {
          const name = 'SM Rot ' + c.word({ length: 8 })
          const description = c.word({ length: 10 })
          const tz = c.pickone(['America/Chicago', 'Africa/Accra', 'Etc/UTC'])
          const shiftLength = c.integer({ min: 1, max: 10 })
          const start = DateTime.fromISO((c.date() as Date).toISOString())

          cy.pageFab()
          cy.dialogTitle('Create Rotation')
          cy.dialogForm({
            name,
            description,
            timeZone: tz,
            type,
            dayOfWeek: type === 'Weekly' ? start.weekdayLong : null,
            shiftLength: shiftLength.toString(),
            start: '15:04',
          })
          cy.dialogFinish('Submit')

          // should be on details page
          cy.get('body')
            .should('contain', name)
            .should('contain', description)
            .should('contain', tz)
        })
      })
    })
  })

  describe('Details Page', () => {
    let rot: Rotation
    beforeEach(() =>
      cy.createRotation({ count: 3 }).then(r => {
        rot = r
        return cy.visit(`/rotations/${r.id}`)
      }),
    )

    it('should display users correctly', () => {
      cy.get('ul[data-cy=users]')
        .find('li')
        .as('parts')
      cy.get('@parts')
        .eq(1)
        .should('contain', rot.users[0].name)
      cy.get('@parts')
        .eq(2)
        .should('contain', rot.users[1].name)
    })

    it('should allow removing a user', () => {
      cy.get('ul[data-cy=users]')
        .find('li')
        .as('parts')

      // remove second user
      cy.get('@parts')
        .eq(2)
        .find('button')
        .menu('Remove')

      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.get('@parts').should('not.contain', rot.users[1].name)

      // add again
      cy.pageFab()
      cy.dialogTitle('Add User')
      cy.dialogForm({ users: rot.users[1].name })
      cy.dialogFinish('Submit')

      cy.get('ul[data-cy=users]')
        .find('li')
        .should('contain', rot.users[1].name)
    })

    it('should allow re-ordering participants', () => {
      // ensure list has fully loaded before drag/drop
      cy.get('ul[data-cy=users]')
        .find('li')
        .should('have.length', 4)
      cy.get('[data-cy=avatar-fallback]').should('not.exist')

      cy.get('ul[data-cy=users]')
        .find('li')
        .eq(1)
        .should('contain', rot.users[0].name)
        .should('contain', 'Active')
        .parent('[tabindex]')
        .focus()
        .type(' ')

      cy.get('body').should('contain', 'You have lifted an item in position 1')

      cy.focused().type('{downarrow}', { force: true })

      cy.get('body')
        .should('contain', 'You have moved the item from position 1')
        .should('contain', 'to position 2')

      cy.focused().type(' ', { force: true })

      cy.get('ul[data-cy=users]')
        .find('li')
        .as('parts')
      cy.get('@parts')
        .eq(1)
        .should('contain', rot.users[1].name)
        .should('not.contain', 'Active')
      cy.get('@parts')
        .eq(2)
        .should('contain', rot.users[0].name)
        .should('contain', 'Active')
      cy.get('@parts')
        .eq(3)
        .should('contain', rot.users[2].name)
        .should('not.contain', 'Active')
    })

    it('should allow changing the active user', () => {
      cy.get('ul[data-cy=users]')
        .find('li')
        .as('parts')

      cy.get('@parts')
        .eq(2)
        .find('button')
        .menu('Set Active')

      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.get('@parts')
        .eq(1)
        .should('contain', rot.users[0].name)
        .should('not.contain', 'Active')
      cy.get('@parts')
        .eq(2)
        .should('contain', rot.users[1].name)
        .should('contain', 'Active')
      cy.get('@parts')
        .eq(3)
        .should('contain', rot.users[2].name)
        .should('not.contain', 'Active')
    })

    it('should allow deleting the rotation', () => {
      cy.pageAction('Delete')

      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.url().should('eq', Cypress.config().baseUrl + '/rotations')
      cy.pageSearch(rot.name)
      cy.get('body').should('contain', 'No results')
    })
  })

  it('should allow editing a rotation', () => {
    cy.createRotation({ shiftLength: 3, type: 'daily' }).then(r => {
      const newName = c.word({ length: 15 })
      const newDesc = c.sentence({ words: 3 })
      const newTz = 'Africa/Accra'

      cy.visit(`/rotations/${r.id}`)
      cy.pageAction('Edit Rotation')

      cy.dialogTitle('Edit Rotation')
      cy.dialogForm({
        name: newName,
        description: newDesc,
        timeZone: newTz,
        type: 'Weekly',
        shiftLength: '5',
      })
      cy.dialogFinish('Submit')

      cy.get('body')
        .should('contain', newName)
        .should('contain', newDesc)
        .should('contain', newTz)
    })
  })
}
