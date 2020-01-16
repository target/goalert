import { Chance } from 'chance'
import { testScreen } from '../support'

const c = new Chance()

testScreen('Calendar Subscriptions', testSubs)

function testSubs(screen: ScreenFormat) {
  describe('Creation', () => {
    let sched: Schedule
    beforeEach(() => {
      cy.createSchedule().then(s => {
        sched = s
      })
    })

    /** DONE **/
    it.only('should create a subscription from a schedule', () => {
      const name = c.word({ length: 5 })

      cy.visit(`/schedules/${sched.id}`)

      cy.get('body').should('contain', 'Subscribe to your shifts')
      cy.get('body').should('not.contain', 'You have 1 active subscription')

      // fill form out and submit
      cy.get('button[data-cy="subscribe-btn"]').click()
      cy.dialogTitle('Create New Calendar Subscription')
      cy.dialogForm({
        name,
        'reminderMinutes[0]': 'At time of shift'
      })
      cy.dialogClick('Submit')
      cy.dialogTitle('Success!')
      // todo: verify url generation
      cy.dialogFinish('Done')

      cy.get('body').should('not.contain', 'Subscribe to your shifts')
      cy.get('body').should('contain', 'You have 1 active subscription')
    })

    /** DONE **/
    it('should create a subscription from their profile', () => {
      const name = c.word({ length: 5 })

      cy.visit('/profile/schedule-calendar-subscriptions')

      cy.get('span[data-cy="empty-message-cptn"]').should('contain', 'You are not subscribed to any schedules.')
      cy.get('body').should('not.contain', name)

      // fill form out and submit
      cy.pageFab()
      cy.dialogTitle('Create New Calendar Subscription')
      cy.dialogForm({
        name,
        scheduleID: sched.name,
        'reminderMinutes[0]': 'At time of shift'
      })
      cy.dialogClick('Submit')
      cy.dialogTitle('Success!')
      // todo: verify url generation
      cy.dialogFinish('Done')

      cy.get('span[data-cy="empty-message-cptn"]').should('not.contain', 'You are not subscribed to any schedules.')
      cy.get('body').should('contain', name)
    })

    it('should add and remove additional valarms', () => {})
  })

  describe('Schedule', () => {
    /** DONE **/
    it('should go to the profile subscriptions list from a schedule', () => {
      const flatListText = 'Showing your current on-call subscriptions for all schedules'
      cy.get('body').should('not.contain', flatListText)
      cy.get('a[data-cy="manage-subscriptions-link"]').should('contain', 'Manage subscriptions').click()
      cy.get('body').should('contain', flatListText)
    })
  })

  describe('Profile', () => {
    // let cs CalendarSubscription
    beforeEach(() => {
      cy.visit('/profile/schedule-calendar-subscriptions')
      // cy.createCalendarSubscription().then(sub => {
      //   cs = sub
      // })
    })

    /** DONE **/
    it('should navigate to and from the subscriptions list', () => {
      cy.visit('/profile')
      cy.navigateToAndFrom(
        screen,
        'Profile',
        'Profile',
        'Schedule Calendar Subscriptions',
        '/profile/schedule-calendar-subscriptions',
      )
    })

    it('should view the subscriptions list', () => {})

    it('should edit a subscription', () => {})

    it('should edit a subscription without generating a new url', () => {})

    it('should delete a subscription', () => {})

    it('should visit a schedule from the subheader link', () => {})
  })
}
