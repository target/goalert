declare namespace Cypress {
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
