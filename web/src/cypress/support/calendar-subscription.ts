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
       * Creates an amount of random calendar subscriptions given
       * by a specified count.
       */
      createManyCalendarSubscriptions: typeof createManyCalendarSubscriptions
    }
  }

  interface CalendarSubscription {
    id: string
    name: string
    reminderMinutes: Array<number>
    scheduleID: string
    lastAccess?: string
    disabled?: boolean
  }

  interface CalendarSubscriptionOptions {
    name?: string
    reminderMinutes?: Array<number>
    scheduleID?: string
    disabled?: boolean
  }
}

/*
 * Generate a random array for the reminderMinutes variable
 */
function chanceReminderMinutes(): Array<number> {
  const len = c.integer({ min: 1, max: 5 })
  let reminderMinutes: Array<number> = []
  for (let i = 0; i < len; i++) {
    reminderMinutes.push(c.integer({ min: 0, max: 1440 }))
  }
  return reminderMinutes
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
    reminderMinutes = chanceReminderMinutes()
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

function createManyCalendarSubscriptions(count: number, scheduleID: string): Cypress.Chainable<Array<CalendarSubscription>> {
  return cy.fixture('profile').then(prof => {
    const userID = prof.id

    // create schedule if no ID is provided
    if (!scheduleID) {
      return cy
        .createSchedule()
        .then(s => createManyCalendarSubscriptions(count, s.id))
    }

    let subs: Array<CalendarSubscription> = []
    for (let i = 0; i < count; i++) {
      subs.push({
        id: c.guid(),
        name: 'SM Subscription ' + c.word({ length: 8 }),
        reminderMinutes: chanceReminderMinutes(),
        scheduleID: scheduleID,
      })
    }

    const dbQuery =
      `insert into user_calendar_subscriptions (id, name, user_id, schedule_id, config) values` +
      subs
        .map(p => `('${p.id}', '${p.name}', '${userID}', '${p.scheduleID}', '${JSON.stringify({ ReminderMinutes: p.reminderMinutes })}')`)
        .join(',') +
      `;`

    return cy.sql(dbQuery).then(() => subs)
  })
}

Cypress.Commands.add('createCalendarSubscription', createCalendarSubscription)
Cypress.Commands.add('createManyCalendarSubscriptions', createManyCalendarSubscriptions)
