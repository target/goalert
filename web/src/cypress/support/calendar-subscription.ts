import { Chance } from 'chance'

const c = new Chance()

declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Creates a new calendar subscription.
       */
      createCalendarSubscription: typeof createCalendarSubscription

      /**
       * Deletes all calendar subscriptions given the specified user ID.
       * Will default to deleting all subscriptions for 'profile' if no ID provided.
       */
      resetCalendarSubscriptions: typeof resetCalendarSubscriptions
    }
  }

  interface CalendarSubscription {
    id: string
    name: string
    reminderMinutes: Array<number>
    scheduleID: string
    schedule: Schedule
    lastAccess: string
    disabled: boolean
  }

  interface CalendarSubscriptionOptions {
    name?: string
    reminderMinutes?: Array<number>
    scheduleID?: string
    schedule?: ScheduleOptions
    disabled?: boolean
  }
}

/*
 * Generate a random array for the reminderMinutes variable
 */
function chanceReminderMinutes(): Array<number> {
  const len = c.integer({ min: 1, max: 5 })
  const reminderMinutes: Array<number> = []
  for (let i = 0; i < len; i++) {
    reminderMinutes.push(c.pickone([0, 5, 10, 30, 60, 1440]))
  }
  return reminderMinutes
}

function createCalendarSubscription(
  cs?: CalendarSubscriptionOptions,
): Cypress.Chainable<CalendarSubscription> {
  const mutation = `
    mutation($input: CreateUserCalendarSubscriptionInput!) {
      createUserCalendarSubscription(input: $input) {
        id
        name
        reminderMinutes
        scheduleID
        schedule {
          name
        }
        lastAccess
        disabled
      }
    }
  `

  // create schedule if no ID is provided
  if (!cs?.scheduleID) {
    return cy
      .createSchedule(cs?.schedule)
      .then(s => createCalendarSubscription({ ...cs, scheduleID: s.id }))
  }

  // create reminderMinutes array if not provided
  let reminderMinutes = cs?.reminderMinutes
  if (!reminderMinutes) {
    reminderMinutes = chanceReminderMinutes()
  }

  // create and return subscription
  return cy
    .graphql2(mutation, {
      input: {
        name: cs?.name || 'SM Subscription ' + c.word({ length: 8 }),
        reminderMinutes,
        scheduleID: cs.scheduleID,
        disabled: cs?.disabled || false,
      },
    })
    .then(res => res.createUserCalendarSubscription)
}

function resetCalendarSubscriptions(userID?: string): Cypress.Chainable<void> {
  if (!userID) {
    return cy.fixture('profile').then(prof => {
      resetCalendarSubscriptions(prof.id)
    })
  }

  const dbQuery = `delete from user_calendar_subscriptions where user_id = '${userID}'`
  return cy.sql(dbQuery)
}

Cypress.Commands.add('createCalendarSubscription', createCalendarSubscription)
Cypress.Commands.add('resetCalendarSubscriptions', resetCalendarSubscriptions)
