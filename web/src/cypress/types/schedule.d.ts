// eslint-disable-next-line @typescript-eslint/no-unused-vars
namespace Cypress {
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
  schedule?: ScheduleOptions

  target?: TargetRotationOptions | Target

  rules?: ScheduleRuleOptions[]
}

interface ScheduleTarget {
  schedule: Schedule
  target: Target
  rules: ScheduleRule[]
}

interface ScheduleRule {
  start: string
  end: string
  weekdayFilter: boolean[]
}

interface ScheduleRuleOptions {
  start?: string
  end?: string
  weekdayFilter?: boolean[]
}

interface Schedule {
  id: string
  name: string
  description: string
  timeZone: string
  isFavorite: boolean
}

interface ScheduleOptions {
  name?: string
  description?: string
  timeZone?: string
  isFavorite?: boolean
}
