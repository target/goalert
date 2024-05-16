import { testScreen } from '../support/e2e'
import { Chance } from 'chance'
import { Schedule } from '../../schema'

const c = new Chance()

function check(
  typeName: string,
  urlPrefix: string,
  createFunc: (name: string, fav: boolean) => Cypress.Chainable<string>,
  getSearchSelectFunc?: () => Cypress.Chainable<JQuery<HTMLElement>>,
  getSearchSelectItemsFunc?: (
    sel: Cypress.Chainable<JQuery<HTMLElement>>,
    prefix: string,
  ) => Cypress.Chainable<JQuery<HTMLElement>>,
): void {
  describe(typeName + ' Favorites', () => {
    it('should allow setting and unsetting as a favorite from details page ', () => {
      createFunc('', false).then((id) => {
        cy.visit(`/${urlPrefix}/${id}`)
        typeName = typeName.toLowerCase()
        // test setting as favorite
        cy.get(`button[aria-label="Set as a Favorite ${typeName}"]`).click()
        // aria label should change and should be set as a favorite, test unsetting

        cy.get(`button[aria-label="Unset as a Favorite ${typeName}"`).click()

        // check that unset
        cy.get(`button[aria-label="Set as a Favorite ${typeName}"]`).click()
      })
    })
    it('should list favorites at the top', () => {
      const prefix = c.word({ length: 12 })
      const name1 = prefix + ' A'
      const name2 = prefix + ' Z'
      createFunc(name1, false)
      createFunc(name2, true)
      cy.visit(`/${urlPrefix}?search=${encodeURIComponent(prefix)}`)

      cy.get('main ul > li')
        .should('have.length', 2)
        .first()
        .should('contain', name2)
        .find('[data-cy=fav-icon]')
        .should('exist')

      cy.get('main ul > li').last().should('contain', name1)
    })
    if (getSearchSelectFunc) {
      it('should sort favorites-first in a search-select', () => {
        const prefix = c.word({ length: 20 })
        const name1 = prefix + ' A'
        const name2 = prefix + ' Z'
        createFunc(name1, false)
        createFunc(name2, true)

        const sel = getSearchSelectFunc()
        const items = getSearchSelectItemsFunc
          ? getSearchSelectItemsFunc(sel, prefix)
          : sel.findByLabel(prefix).get('[data-cy=search-select-item]')

        items.should('have.length.within', 2, 3).as('items') // single selects include value

        cy.get('@items').should('contain', name2)
        cy.get('@items').should('contain', name1)
      })
    }
  })
}

function testFavorites(screen: ScreenFormat): void {
  check(
    'Service',
    'services',
    (name: string, favorite: boolean) =>
      cy.createService({ name, favorite }).then((s: Service) => s.id),
    () => {
      const summary = c.sentence({
        words: 3,
      })

      cy.visit('/alerts')

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Alert').click()
      }
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
      cy.createRotation({ name, favorite }).then((r: Rotation) => r.id),
    () => {
      cy.createEP().then((e: EP) => {
        return cy.visit(`/escalation-policies/${e.id}`)
      })

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Step').click()
      }

      cy.dialogForm({ 'dest.type': 'Rotation' })

      return cy.get('input[name=rotation-id]')
    },
  )

  check(
    'Schedule',
    'schedules',
    (name: string, isFavorite: boolean) =>
      cy
        .createSchedule({ name, isFavorite })
        .then((sched: Schedule) => sched.id),
    () => {
      cy.createEP().then((e: EP) => {
        return cy.visit(`/escalation-policies/${e.id}`)
      })
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Step').click()
      }

      cy.dialogForm({ 'dest.type': 'Schedule' })
      return cy.get('input[name=schedule-id]')
    },
  )

  check(
    'Escalation Policy',
    'escalation-policies',
    (name: string, favorite: boolean) =>
      cy.createEP({ name, favorite }).then((ep: EP) => ep.id),
    () => {
      cy.createService().then((service: Service) => {
        return cy.visit(`/services/${service.id}`)
      })
      cy.get('button[aria-label=Edit]').click()
      return cy.get('input[name=escalation-policy]')
    },
  )
  check(
    'User',
    'users',
    (name: string, favorite: boolean) =>
      cy.createUser({ name, favorite }).then((user: Profile) => user.id),
    () => {
      cy.createEP().then((e: EP) => {
        return cy.visit(`/escalation-policies/${e.id}`)
      })
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Step').click()
      }

      cy.dialogForm({ 'dest.type': 'User' })
      return cy.get('input[name=user-id]')
    },
  )
}

testScreen('Favorites', testFavorites)
