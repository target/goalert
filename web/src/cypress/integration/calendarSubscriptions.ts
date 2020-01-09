import { Chance } from 'chance'
import { testScreen } from '../support'

const c = new Chance()

testScreen('Calendar Subscriptions', testSubs)

function testSubs(screen: ScreenFormat) {
  describe('Schedule', () => {
    it('should create a subscription from a schedule', () => {})

    it('should go to the profile subscriptions list from a schedule', () => {})
  })

  describe('Profile', () => {
    // let cs CalendarSubscription
    beforeEach(() => {
      cy.visit('/profile/schedule-calendar-subscriptions')
      // cy.createCalendarSubscription().then(sub => {
      //   cs = sub
      // })
    })

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

    it('should create a subscription from their profile', () => {})

    it('should edit a subscription', () => {})

    it('should edit a subscription without generating a new url', () => {})
  })
}
