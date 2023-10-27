import { testScreen } from '../support/e2e'
import { Chance } from 'chance'
import prof from '../fixtures/profile.json'

const c = new Chance()

function testUsers(screen: ScreenFormat): void {
  describe('List Page', () => {
    let cm: ContactMethod
    beforeEach(() => {
      cy.addContactMethod({ type: 'SMS' })
        .then((_cm: ContactMethod) => {
          cm = _cm
        })
        .visit('/users')
    })

    it('should handle searching', () => {
      cy.get('ul[data-cy=paginated-list]').should('exist')
      // by name
      cy.pageSearch(prof.name)
      // cypress user and cypress admin
      cy.get('[data-cy=paginated-list] > li').should('have.lengthOf', 2)
      cy.get('ul').should('contain', prof.name)
    })

    it('should handle searching by phone number', () => {
      if (screen === 'mobile') {
        cy.get('[data-cy=app-bar] button[data-cy=open-search]').click()
      }
      cy.get('button[data-cy="users-filter-button"]').click()
      cy.form({ 'user-phone-search': cm.value })
      cy.get('[data-cy=paginated-list] > li').should('have.lengthOf', 1)
      cy.get('ul').should('contain', prof.name)
    })
  })

  describe('Details Page', () => {
    let user: Profile
    beforeEach(() =>
      cy.createUser().then((u: Profile) => {
        user = u
        cy.adminLogin()
        return cy.visit(`/users/${user.id}`)
      }),
    )
    it('should display correct information', () => {
      cy.get('body').should('contain', user.name).should('contain', user.email)
    })

    it('should edit a user role', () => {
      cy.get('[data-cy="card-actions"]')
        .find('button[aria-label="Edit Access"]')
        .click()
      cy.get('[type="checkbox"]').check()
      cy.dialogFinish('Submit')

      cy.reload()
      cy.get('[data-cy="card-actions"]')
        .find('button[aria-label="Edit Access"]')
        .click()
      cy.get('[type="checkbox"]').should('be.checked')
    })

    it('should delete a user', () => {
      cy.get('[data-cy="card-actions"]')
        .find('button[aria-label="Delete"]')
        .click()
      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')
      cy.get('[data-cy=paginated-list]').should('not.contain', user.name)
    })

    describe('User Password', () => {
      beforeEach(() => {
        cy.get('[data-cy="card-actions"]')
          .find('button[aria-label="Edit Access"]')
          .click()
      })
      it('should show error when username is missing', () => {
        cy.get('input[name="password"]').type('test')
        cy.get('input[name="confirmNewPassword"]').type('test')
        cy.dialogClick('Submit')
        cy.get('input[name="username"]')
          .parent()
          .parent()
          .next('p')
          .should('contain', 'Username required')
      })

      it('should show error when password length is too short', () => {
        cy.get('input[name="username"]').type('test')
        cy.get('input[name="password"]').type('test')
        cy.get('input[name="confirmNewPassword"]').type('test')
        cy.dialogClick('Submit')
        cy.get('input[name="password"]')
          .parent()
          .parent()
          .next('p')
          .should('contain', 'Must be at least 8 characters')
      })

      it('should show error when passwords do not match', () => {
        cy.get('input[name="password"]').type('example123')
        cy.get('input[name="confirmNewPassword"]').type('example456')
        cy.dialogClick('Submit')
        cy.get('input[name="confirmNewPassword"]')
          .parent()
          .parent()
          .next('p')
          .should('contain', 'Passwords do not match')
      })

      it("should handle resetting a user's password as an admin", () => {
        cy.get('input[name="password"]').type('ValidPassword')
        cy.get('input[name="confirmNewPassword"]').type('ValidPassword')
        cy.dialogClick('Submit')
      })
    })
  })

  describe('User Subpages', () => {
    it('should navigate to and from its on-call assignments', () => {
      cy.createUser().then((user: Profile) => {
        cy.visit(`users/${user.id}`)

        cy.navigateToAndFrom(
          screen,
          'Users',
          user.name,
          'On-Call Assignments',
          `${user.id}/on-call-assignments`,
        )
      })
    })

    it('should see no on-call assignments text', () => {
      cy.createUser().then((user: Profile) => {
        cy.visit(`users/${user.id}`)

        cy.get('[data-cy=route-links]').contains('On-Call Assignments').click()
        cy.get('body').should(
          'contain',
          `${user.name} is not currently on-call.`,
        )
      })
    })

    it('should see on-call assignment list', () => {
      const name = 'SVC ' + c.word({ length: 8 })
      cy.createUser().then((user: Profile) => {
        cy.visit(`users/${user.id}`)

        return cy
          .createService({ name })
          .then((svc: Service) => {
            return cy
              .createEPStep({
                epID: svc.epID,
                targets: [{ type: 'user', id: user.id }],
              })
              .engineTrigger()
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
      cy.createUser().then((user: Profile) => {
        cy.adminLogin()
        cy.visit(`users/${user.id}`)

        cy.navigateToAndFrom(
          screen,
          'Users',
          user.name,
          'Sessions',
          `${user.id}/active-sessions`,
        )
      })
    })

    it('should view and interact with the profile calendar', () => {
      cy.createUser().then((user: Profile) => {
        cy.setScheduleTarget({
          target: {
            id: user.id,
            type: 'user',
          },
        }).then((sched) => {
          cy.visit(`users/${user.id}`)

          cy.get('[data-cy-spin-loading=false]').should('exist')

          cy.get('div').contains(sched.name).click()
          cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
          cy.get(
            'div[data-cy="shift-tooltip"] a:contains("Visit Schedule")',
          ).click()
          cy.url().should('include', sched.id)
        })
      })
    })
  })
}

testScreen('Users', testUsers)
