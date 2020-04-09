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

interface TargetRotationOptions {
  rotation: RotationOptions
}
interface Target {
  type: 'user' | 'rotation'
  id: string
}

interface ScheduleTargetOptions {
  scheduleID?: string
  schedule?: Partial<Schedule>

  target?: TargetRotationOptions | Target

  rules?: Partial<ScheduleRule>[]
}
