declare namespace Cypress {
  interface Chainable {
    /** Creates a new schedule. */
    createSchedule: typeof createSchedule

    /** Deletes a schedule with its specified ID */
    deleteSchedule: typeof deleteSchedule

    /** Configures a schedule target and rules. */
    setScheduleTarget: typeof setScheduleTarget
  }
}
