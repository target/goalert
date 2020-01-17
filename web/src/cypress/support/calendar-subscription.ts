import { Chance } from 'chance'

const c = new Chance()

declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Creates a new calendar subscription.
       */
      createCalendarSubscription: typeof createCalendarSubscription

      /** Delete the calendar subscription with the specified ID */
      // deleteCalendarSubscription: typeof deleteCalendarSubscription
    }
  }

  interface CalendarSubscription {
    id: string
    name: string
    reminderMinutes: Array<number>
    scheduleID: string
    lastAccess: string
    disabled: boolean
  }

  interface CalendarSubscriptionOptions {
    name?: string
    reminderMinutes?: Array<number>
    scheduleID?: string
    disabled?: boolean
  }
}

function createCalendarSubscription(cs?: CalendarSubscriptionOptions): Cypress.Chainable<CalendarSubscription> {
  const mutation = `
    mutation($input: CreateUserCalendarSubscriptionInput!) {
      createUserCalendarSubscription(input: $input) {
        id
        name
        reminderMinutes
        scheduleID
        lastAccess
        disabled
      }
    }
  `

  // create schedule if no ID is provided
  if (!cs?.scheduleID) {
    return cy
      .createSchedule()
      .then(s => createCalendarSubscription({ ...cs, scheduleID: s.id }))
  }

  // create reminderMinutes array if not provided
  let reminderMinutes = cs?.reminderMinutes
  if (!reminderMinutes) {
    const len = c.integer({ min: 1, max: 5 })
    reminderMinutes = []
    for (let i = 0; i < len; i++) {
      reminderMinutes.push(c.integer({ min: 0, max: 1440 }))
    }
  }

  // create and return subscription
  return cy.graphql2(mutation, {
    input: {
      name: cs?.name || 'SM Subscription ' + c.word({ length: 8 }),
      reminderMinutes,
      scheduleID: cs.scheduleID,
      disabled: false,
    }
  }).then(res => res.createUserCalendarSubscription)
}

Cypress.Commands.add('createCalendarSubscription', createCalendarSubscription)
