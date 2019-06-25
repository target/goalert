import { testScreen } from '../support'
import { Chance } from 'chance'
const c = new Chance()

testScreen('Favorites', testFavorites)

function testFavorites(screen: ScreenFormat) {
  // describe('Rotation Favorites', () => {
  //   let rot: Rotation
  //   beforeEach(() => {
  //     cy.createRotation()
  //       .then(r => {
  //         rot = r
  //       })
  //       .visit('/rotations')
  //   })
  detailsCheck('Service', 'services', () => cy.createService().then(s => s.id))
  detailsCheck('Rotation', 'rotations', () =>
    cy.createRotation().then(r => r.id),
  )
}

//     it('should allow setting and unsetting as a favorite rotation', () => {
//       cy.pageSearch(rot.name)
//       cy.get('#app')
//         .contains(rot.name)
//         .click()
//       // test setting as favorite
//       cy.get('button[aria-label="Set as a Favorite Rotation"]').click()
//       cy.reload()
//       // aria label should change and should be set as a favorite, test unsetting
//       cy.get('button[aria-label="Unset as a Favorite Rotation"').click()
//       cy.reload()
//       // check that unset
//       cy.get('button[aria-label="Set as a Favorite Rotation"]').click()
//     })
//
//     it('should have favorited rotations move to first on rotation list', () => {
//       cy.pageSearch(rot.name)
//       cy.get('#app')
//         .contains(rot.name)
//         .click()
//       cy.get('button[aria-label="Set as a Favorite Rotation"]')
//         .click()
//         .visit('/rotations')
//       cy.get('#app')
//         .contains(rot.name)
//         .siblings()
//         .find('svg[data-cy=fav-icon]')
//     })
//
//     it('should have favorites first in RotationSelect on escalation policy steps', () => {
//       cy.pageSearch(rot.name)
//       cy.get('#app')
//         .contains(rot.name)
//         .click()
//       cy.get('button[aria-label="Set as a Favorite Rotation"]')
//         .click()
//         .visit('/escalation-policies')
//
//       let ep = null
//       cy.createEP().then(e => {
//         ep = e
//         cy.visit(`/escalation-policies/${ep.id}`)
//         cy.reload()
//       })
//
//       cy.get('button[data-cy=page-fab]').click()
//
//       cy.get('[data-cy=search-select-input]').click()
//       cy.get('[data-cy=select-dropdown]').contains(rot.name)
//     })
//
//     it('should show favorites first in schedule assignment', () => {
//       cy.pageSearch(rot.name)
//       cy.get('#app')
//         .contains(rot.name)
//         .click()
//       cy.get('button[aria-label="Set as a Favorite Rotation"]').click()
//
//       let sch = null
//       cy.createSchedule().then(s => {
//         sch = s
//         cy.visit(`/schedules/${sch.id}`)
//         cy.reload()
//       })
//
//       cy.get('a[role=button]')
//         .eq(0)
//         .click()
//
//       cy.pageFab('Rotation')
//       cy.get('input[name=targetID]')
//
//       cy.get('[data-cy=search-select-input]').click()
//       cy.get('[data-cy=select-dropdown]').contains(rot.name)
//     })
//   })
//
//   describe('Service Favorites', () => {
//     let svc: Service
//     beforeEach(() =>
//       cy.createService().then(s => {
//         svc = s
//         return cy.visit(`/services/${s.id}`)
//       }),
//     )
//
//     it('should allow setting and unsetting as a favorite service', () => {
//       // test setting as favorite
//       cy.get('button[aria-label="Set as a Favorite Service"]').click()
//       cy.reload()
//       // aria label should change and should be set as a favorite, test unsetting
//       cy.get('button[aria-label="Unset as a Favorite Service"').click()
//       cy.reload()
//       // check that unset
//       cy.get('button[aria-label="Set as a Favorite Service"]').click()
//     })
//   })
// }

function detailsCheck(typeName: string, urlPrefix: string, createFunc: any) {
  describe(typeName + ' Favorites', () => {
    beforeEach(() =>
      createFunc().then((id: string) => cy.visit(`/${urlPrefix}/${id}`)),
    )

    it('should allow setting and unsetting as a favorite ' + typeName, () => {
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
}

function listsCheck(typeName: string, urlPrefix: string, createFunc: any) {
  describe(typeName + 'List Favorites', () => {
    beforeEach(() =>
      createFunc().then((id: string) => cy.visit(`/${urlPrefix}/${id}`)),
    )
  })
}
