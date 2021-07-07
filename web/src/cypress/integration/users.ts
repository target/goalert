import { testScreen } from '../support'
import { User } from '../../schema'
import { Chance } from 'chance'

const c = new Chance()

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

  describe('Details Page', () => {
    let user: User
    beforeEach(() =>
      cy.createUser().then((u: User) => {
        user = u
        cy.adminLogin()
        return cy.visit(`/users/${user.id}`)
      }),
    )

    it('should display correct information', () => {
      cy.get('body').should('contain', user.name).should('contain', user.email)
    })

    it('should edit a user role', () => {
      cy.get('[data-cy="card-actions"]').find('button[title="Edit"]').click()
      cy.get('[type="checkbox"]').check()
      cy.dialogFinish('Confirm')

      cy.reload()
      cy.get('[data-cy="card-actions"]').find('button[title="Edit"]').click()
      cy.get('[type="checkbox"]').should('be.checked')
    })

    it('should delete a user', () => {
      cy.get('[data-cy="card-actions"]').find('button[title="Delete"]').click()
      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.get('[data-cy=apollo-list]').should('not.contain', user.name)
    })
  })

  describe('User Subpages', () => {
    it('should navigate to and from its on-call assignments', () => {
      cy.createUser().then((user: User) => {
        cy.visit(`users/${user.id}`)

        cy.navigateToAndFrom(
          screen,
          'User Details',
          user.name,
          'On-Call Assignments',
          `${user.id}/on-call-assignments`,
        )
      })
    })

    it('should see no on-call assignments text', () => {
      cy.createUser().then((user: User) => {
        cy.visit(`users/${user.id}`)

        cy.get('[data-cy=route-links]').contains('On-Call Assignments').click()
        cy.get('body').should(
          'contain',
          `${user.name} is not currently on-call.`,
        )
      })
    })

    it('should see on-call assigment list', () => {
      const name = 'SVC ' + c.word({ length: 8 })
      cy.createUser().then((user: User) => {
        cy.visit(`users/${user.id}`)

        return cy
          .createService({ name })
          .then((svc: Service) => {
            return cy
              .fixture('users')
              .then(() => {
                return cy.createEPStep({
                  epID: svc.epID,
                  targets: [{ type: 'user', id: user.id }],
                })
              })
              .task('engine:trigger')
              .then(() => svc.id)
          })
          .then((svcID: string) => {
            cy.get('[data-cy=route-links]')
              .contains('On-Call Assignments')
              .click()
            cy.get('body').contains('a', name).click()
            cy.url().should(
              'eq',
              Cypress.config().baseUrl + '/services/' + svcID,
            )
          })
      })
    })

    // admin only
    it('should navigate to and from its active sessions', () => {
      cy.createUser().then((user: User) => {
        cy.adminLogin()
        cy.visit(`users/${user.id}`)

        cy.navigateToAndFrom(
          screen,
          'User Details',
          user.name,
          'Sessions',
          `${user.id}/sessions`,
        )
      })
    })
  })
}

testScreen('Users', testUsers)
