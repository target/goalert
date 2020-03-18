import { testScreen } from '../support'
import { Chance } from 'chance'
const c = new Chance()

testScreen('Favorites', testFavorites)

function testFavorites() {
  check(
    'Service',
    'services',
    (name: string, favorite: boolean) =>
      cy.createService({ name, favorite }).then(s => s.id),
    () => {
      const summary = c.sentence({
        words: 3,
      })

      cy.visit('/alerts')

      cy.pageFab()
      cy.dialogTitle('New Alert')
      cy.dialogForm({ summary })
      cy.dialogClick('Next')

      return cy.get('input[name=serviceSearch]')
    },
    (sel: Cypress.Chainable<JQuery<HTMLElement>>, prefix: string) =>
      sel
        .type(prefix)
        .get('ul[data-cy=service-select] [data-cy=service-select-item]'),
  )

  check(
    'Rotation',
    'rotations',
    (name: string, favorite: boolean) =>
      cy.createRotation({ name, favorite }).then(r => r.id),
    () =>
      cy
        .createEP()
        .then(e => {
          return cy.visit(`/escalation-policies/${e.id}`)
        })
        .pageFab()
        .get('input[name=rotations]'),
  )

  check(
    'Schedule',
    'schedules',
    (name: string, isFavorite: boolean) =>
      cy.createSchedule({ name, isFavorite }).then(sched => sched.id),
    () =>
      cy
        .createEP()
        .then(e => {
          return cy.visit(`/escalation-policies/${e.id}`)
        })
        .pageFab()
        .get('[data-cy="schedules-step"]')
        .click()
        .get('input[name=schedules]'),
  )
}

function check(
  typeName: string,
  urlPrefix: string,
  createFunc: (name: string, fav: boolean) => Cypress.Chainable<string>,
  getSearchSelectFunc?: () => Cypress.Chainable<JQuery<HTMLElement>>,
  getSearchSelectItemsFunc?: (
    sel: Cypress.Chainable<JQuery<HTMLElement>>,
    prefix: string,
  ) => Cypress.Chainable<JQuery<HTMLElement>>,
) {
  describe(typeName + ' Favorites', () => {
    it('should allow setting and unsetting as a favorite from details page ', () => {
      createFunc('', false).then(id => {
        cy.visit(`/${urlPrefix}/${id}`)
        typeName = typeName.toLowerCase()
        // test setting as favorite
        cy.get(`button[aria-label="Set as a Favorite ${typeName}"]`).click()
        cy.reload()
        // aria label should change and should be set as a favorite, test unsetting
        cy.get(`button[aria-label="Unset as a Favorite ${typeName}"`).click()
        cy.reload()
        // check that unset
        cy.get(`button[aria-label="Set as a Favorite ${typeName}"]`).click()
      })
    })
    it('should list favorites at the top', () => {
      const prefix = c.word({ length: 12 })
      const name1 = prefix + 'A'
      const name2 = prefix + 'Z'
      createFunc(name1, false)
      createFunc(name2, true)
      cy.visit(`/${urlPrefix}?search=${encodeURIComponent(prefix)}`)

      cy.get('ul[data-cy=apollo-list] li')
        .should('have.length', 2)
        .first()
        .should('contain', name2)
        .find('[data-cy=fav-icon]')
        .should('exist')

      cy.get('ul[data-cy=apollo-list] li')
        .last()
        .should('contain', name1)
    })
    if (getSearchSelectFunc) {
      it('should sort favorites-first in a search-select', () => {
        const prefix = c.word({ length: 12 })
        const name1 = prefix + 'A'
        const name2 = prefix + 'Z'
        createFunc(name1, false)
        createFunc(name2, true)

        const sel = getSearchSelectFunc()
        const items = getSearchSelectItemsFunc
          ? getSearchSelectItemsFunc(sel, prefix)
          : sel
              .findByLabel(prefix)
              .parent()
              .children()

        items.should('have.length', 2).as('items')

        cy.get('@items')
          .first()
          .should('contain', name2)
        cy.get('@items')
          .last()
          .should('contain', name1)
      })
    }
  })
}
