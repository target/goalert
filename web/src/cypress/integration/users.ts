import { testScreen } from '../support'

function testUsers(screen: ScreenFormat): void {
  describe('List Page', () => {
    let cm: ContactMethod
    let prof: Profile
    beforeEach(() => {
      cy.addContactMethod({ type: 'SMS' })
        .then((_cm: ContactMethod) => {
          cm = _cm
          return cy.fixture('profile').then((_prof: Profile) => {
            prof = _prof
          })
        })
        .visit('/users')
    })

    it('should handle searching', () => {
      cy.get('ul[data-cy=apollo-list]').should('exist')
      // by name
      cy.pageSearch(prof.name)
      // cypress user and cypress admin
      cy.get('[data-cy=apollo-list] > li').should('have.lengthOf', 2)
      cy.get('ul').should('contain', prof.name)
    })

    it('should handle searching by phone number', () => {
      if (screen === 'mobile') {
        cy.get('[data-cy=app-bar] button[data-cy=open-search]').click()
      }
      cy.get('button[data-cy="users-filter-button"]').click()
      cy.form({ 'user-phone-search': cm.value })
      cy.get('[data-cy=apollo-list] > li').should('have.lengthOf', 1)
      cy.get('ul').should('contain', prof.name)
    })
  })
}

testScreen('Users', testUsers)
