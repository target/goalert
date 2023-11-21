import { Chance } from 'chance'
import { testScreen } from '../support/e2e'
import { Schedule } from '../../schema'
import users from '../fixtures/users.json'
import profile from '../fixtures/profile.json'

const c = new Chance()

function testSubs(screen: ScreenFormat): void {
  beforeEach(() => {
    cy.resetCalendarSubscriptions()
  })

  describe('Creation', () => {
    let sched: Schedule
    beforeEach(() => {
      cy.createSchedule().then((s: Schedule) => {
        sched = s
      })
    })

    it('should create a subscription from a schedule', () => {
      const name = c.word({ length: 5 })
      const defaultCptn =
        'Subscribe to your personal shifts from your preferred calendar app'

      cy.visit(`/schedules/${sched.id}`)

      cy.get('[data-cy="subscribe-btn"]').trigger('mouseover')
      cy.get('[data-cy="subscribe-btn-context"]').should('contain', defaultCptn)
      cy.get('body').should('not.contain', 'You have 1 active subscription')

      // fill form out and submit
      cy.get('[data-cy="subscribe-btn"]').click()
      cy.dialogTitle('Create New Calendar Subscription')
      cy.dialogForm({ name })
      cy.dialogClick('Submit')
      cy.dialogTitle('Success!')
      cy.dialogContains(Cypress.config().baseUrl + '/api/v2/calendar?token=')
      cy.dialogFinish('Done')

      cy.get('body').should('not.contain', defaultCptn)
      cy.get('body').should('contain', 'You have 1 active subscription')
    })

    it('should create a subscription from their profile', () => {
      const name = c.word({ length: 5 })

      cy.visit('/profile/external-calendar-subscriptions')

      cy.get('[data-cy="list-empty-message"]').should(
        'contain',
        'You are not subscribed to any schedules.',
      )
      cy.get('[data-cy=calendar-subscriptions]').should('not.contain', name)

      // fill form out and submit
      cy.pageFab()
      cy.dialogTitle('Create New Calendar Subscription')
      cy.dialogForm({
        name,
        scheduleID: sched.name,
      })
      cy.dialogClick('Submit')
      cy.dialogTitle('Success!')
      cy.dialogContains(Cypress.config().baseUrl + '/api/v2/calendar?token=')
      cy.dialogFinish('Done')

      cy.get('[data-cy="list-empty-message"]').should('not.exist')
      cy.get('[data-cy=calendar-subscriptions]').should('contain', name)
    })
  })

  describe('Schedule', () => {
    let sched: Schedule
    beforeEach(() => {
      cy.createSchedule().then((s: Schedule) => {
        sched = s
        cy.visit(`/schedules/${s.id}`)
      })
    })

    it('should update button caption text after a subscription is created', () => {
      const defaultCptn =
        'Subscribe to your personal shifts from your preferred calendar app'
      const oneSubCptn = 'You have 1 active subscription for this schedule'
      const multipleSubsCptn =
        'You have 2 active subscriptions for this schedule'

      const subsribeBtn = '[data-cy="subscribe-btn"]'
      const context = '[data-cy="subscribe-btn-context"]'

      cy.get(subsribeBtn).trigger('mouseover')
      cy.get(context)
        .should('contain', defaultCptn)
        .should('not.contain', oneSubCptn)
        .should('not.contain', multipleSubsCptn)

      cy.createCalendarSubscription({ scheduleID: sched.id })
      cy.refetchAll()

      cy.get(subsribeBtn).trigger('mouseover')
      cy.get(context)
        .should('not.contain', defaultCptn)
        .should('contain', oneSubCptn)
        .should('not.contain', multipleSubsCptn)

      cy.createCalendarSubscription({ scheduleID: sched.id })
      cy.refetchAll()

      cy.get(subsribeBtn).trigger('mouseover')
      cy.get(context)
        .should('not.contain', defaultCptn)
        .should('not.contain', oneSubCptn)
        .should('contain', multipleSubsCptn)
    })
  })

  describe('Profile', () => {
    let cs: CalendarSubscription
    beforeEach(() => {
      cy.createCalendarSubscription().then((sub: CalendarSubscription) => {
        cs = sub
        cy.visit('/profile/external-calendar-subscriptions')
      })
    })

    it('should navigate to and from the subscriptions list', () => {
      cy.visit('/profile')
      cy.navigateToAndFrom(
        screen,
        'Users',
        'Cypress User',
        'External Calendar Subscriptions',
        `/users/${profile.id}/external-calendar-subscriptions`,
      )
    })

    it('should view the subscriptions list', () => {
      cy.get('body').should(
        'contain',
        'Showing your current on-call subscriptions for all schedules',
      )
      cy.get('[data-cy=calendar-subscriptions]').should('contain', cs.name)
      cy.get('[data-cy=calendar-subscriptions]').should(
        'contain',
        'Last sync: Never',
      )
    })

    it('should edit a subscription', () => {
      const name = 'SM Subscription ' + c.word({ length: 8 })

      cy.get('[data-cy=calendar-subscriptions]').should('contain', cs.name)
      cy.get('[data-cy=calendar-subscriptions]').should(
        'contain',
        'Last sync: Never',
      )

      cy.get('[data-cy=calendar-subscriptions]')
        .contains('li', cs.name)
        .find('[data-cy=other-actions]')
        .menu('Edit')

      cy.dialogTitle('Edit Calendar Subscription')
      cy.dialogForm({ name })
      cy.dialogFinish('Submit')

      cy.get('[data-cy=calendar-subscriptions]').should('contain', name)
      cy.get('[data-cy=calendar-subscriptions]').should(
        'contain',
        'Last sync: Never',
      )
    })

    it('should delete a subscription', () => {
      cy.get('[data-cy="list-empty-message"]').should('not.exist')
      cy.get('[data-cy=calendar-subscriptions]').should('contain', cs.name)

      cy.get('[data-cy=calendar-subscriptions]')
        .contains('li', cs.name)
        .find('[data-cy=other-actions]')
        .menu('Delete')

      cy.dialogFinish('Confirm')

      cy.get('[data-cy="list-empty-message"]').should(
        'contain',
        'You are not subscribed to any schedules.',
      )
      cy.get('[data-cy=calendar-subscriptions]').should('not.contain', cs.name)
    })

    it('should visit a schedule from the subheader link', () => {
      cy.get('[data-cy="subscribe-btn"]').should('not.exist')

      cy.get('[data-cy=calendar-subscriptions]')
        .contains('li a', cs.schedule.name)
        .click()

      cy.get('[data-cy="subscribe-btn"]').should('exist')
    })

    it('should not show route link unless on personal profile', () => {
      cy.visit(`/users/${users[0].id}`)
      cy.get('[data-cy="route-links"]').should(
        'not.contain',
        'External Calendar Subscriptions',
      )
    })

    it('should show an icon if a subscription is disabled, and vice-versa', () => {
      cy.createCalendarSubscription({ disabled: true }).then(
        (disabledCs: CalendarSubscription) => {
          cy.reload()

          cy.get('[data-cy=calendar-subscriptions] li')
            .contains(cs.name)
            .find('[data-cy="warning-icon"]')
            .should('not.exist')

          cy.get('[data-cy=calendar-subscriptions] li')
            .contains(disabledCs.name)
            .parent()
            .parent()
            .find('[data-cy="warning-icon"]') // two divs of separation
            .should('exist')
        },
      )
    })

    it('should show warning message when disabled in config', () => {
      cy.setConfig({
        General: {
          DisableCalendarSubscriptions: true,
        },
      }).then(() => {
        cy.reload()

        cy.get('[data-cy="subs-disabled-warning"]').should('exist')
        cy.get('[data-cy="subs-disabled-warning"]').should(
          'contain',
          'Calendar subscriptions are currently disabled by your administrator',
        )
      })
    })
  })
}

testScreen('Calendar Subscriptions', testSubs)
