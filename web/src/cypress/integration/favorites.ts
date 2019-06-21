import { testScreen } from '../support'
import { Chance } from 'chance'
const c = new Chance()

testScreen('Favorites', testFavorites)

function testFavorites(screen: ScreenFormat) {
  describe('Rotation Favorites', () => {
    let rot: Rotation
    beforeEach(() => {
      cy.createRotation()
        .then(r => {
          rot = r
        })
        .visit('/rotations')
    })

    it('should have favorited rotations move to first on rotation list', () => {
      cy.pageSearch(rot.name)
      cy.get('#app')
        .contains(rot.name)
        .click()
      cy.get('button[aria-label="Set as a Favorite Rotation"]')
        .click()
        .visit('/rotations')
      cy.get('#app').contains(rot.name)
    })

    it('should have favorites first in RotationSelect on escalation policy steps', () => {
      cy.pageSearch(rot.name)
      cy.get('#app')
        .contains(rot.name)
        .click()
      cy.get('button[aria-label="Set as a Favorite Rotation"]')
        .click()
        .visit('/escalation-policies')

      cy.pageFab()

      cy.get('div[role=dialog]').as('dialog')
      cy.get('@dialog').should('contain', 'Create Escalation Policy')

      const name = 'SM EP ' + c.word({ length: 8 })
      const description = c.word({ length: 10 })
      const repeat = c.integer({ min: 0, max: 5 }).toString()

      cy.get('@dialog')
        .find('input[name=name]')
        .type(name)

      cy.get('@dialog')
        .find('textarea[name=description]')
        .type(description)

      cy.get('@dialog')
        .find('input[name=repeat]')
        .selectByLabel(repeat)

      cy.get('@dialog')
        .contains('button', 'Submit')
        .click()

      // should be on details page
      cy.get('body')
        .should('contain', name)
        .should('contain', description)

      cy.get('button[data-cy=page-fab]').click()

      cy.get('[data-cy=search-select-input]').click()
      cy.get('[data-cy=select-dropdown]').contains(rot.name)
    })

    it('should show favorites first in schedule assignment', () => {
      cy.pageSearch(rot.name)
      cy.get('#app')
        .contains(rot.name)
        .click()
      cy.get('button[aria-label="Set as a Favorite Rotation"]')
        .click()
        .visit('/schedules')

      const name = c.word({ length: 8 })
      const description = c.sentence({ words: 5 })

      cy.pageFab()
      cy.get('input[name=name]').type(name)

      cy.get('textarea[name=description]')
        .clear()
        .type(description)
      cy.get('button')
        .contains('Submit')
        .click()

      cy.get('a[role=button]')
        .eq(0)
        .click()

      cy.pageFab('Rotation')
      cy.get('input[name=targetID]')

      cy.get('[data-cy=search-select-input]').click()
      cy.get('[data-cy=select-dropdown]').contains(rot.name)
    })
  })
}
