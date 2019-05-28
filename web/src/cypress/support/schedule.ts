import { Chance } from 'chance'
const c = new Chance()

declare global {
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
  }

  interface ScheduleOptions {
    name?: string
    description?: string
    timeZone?: string
  }
}

function setScheduleTarget(
  tgt?: ScheduleTargetOptions,
): Cypress.Chainable<ScheduleTarget> {
  if (!tgt) {
    tgt = {}
  }
  if (!tgt.scheduleID) {
    return cy
      .createSchedule(tgt.schedule)
      .then(sched => setScheduleTarget({ ...tgt, scheduleID: sched.id }))
  }
  if (!tgt.target) {
    tgt.target = { rotation: {} }
  }
  const rotation = (<TargetRotationOptions>tgt.target).rotation
  if (rotation) {
    return cy
      .createRotation(rotation)
      .then(r =>
        setScheduleTarget({ ...tgt, target: { type: 'rotation', id: r.id } }),
      )
  }
  if (!tgt.rules) {
    tgt.rules = [{}]
  }
  tgt.rules = tgt.rules.map(r => ({
    start: r.start || randClock(),
    end: r.end || randClock(),
    weekdayFilter: r.weekdayFilter || [
      c.bool(),
      c.bool(),
      c.bool(),
      c.bool(),
      c.bool(),
      c.bool(),
      c.bool(),
    ],
  }))

  const mutation = `mutation($input: ScheduleTargetInput!) {updateScheduleTarget(input: $input)}`
  const query = `query($id: ID!, $tgt: TargetInput!){
    schedule(id: $id) {
      id
      name
      description
      timeZone
      target(input: $tgt) {
        target {id, name, type}
        rules {
          start
          end
          weekdayFilter
        }
      }
    }
  }`
  const { scheduleID, target } = tgt
  return cy
    .graphql2(mutation, {
      input: tgt,
    })
    .then(() => {
      return cy
        .graphql2(query, {
          id: scheduleID,
          tgt: target,
        })
        .then(res => {
          const { target, ...schedule } = res.schedule
          return {
            ...target,
            schedule,
          }
        })
    })
}

function createSchedule(sched?: ScheduleOptions): Cypress.Chainable<Schedule> {
  const query = `mutation createSchedule($input: CreateAllInput!){
          createAll(input: $input) {
            schedules {
              id
              name
              description
              timeZone: time_zone
            }
          }
      }`

  if (!sched) sched = {}

  return cy
    .graphql(query, {
      input: {
        schedules: [
          {
            name: sched.name || 'SM Sched ' + c.word({ length: 8 }),
            description: sched.description || c.sentence(),
            time_zone: sched.timeZone || 'America/Chicago',
          },
        ],
      },
    })
    .then(res => res.createAll.schedules[0])
}

const fmtTime = (str: any) => {
  const s = str.toString()
  if (s.length === 1) {
    return '0' + s
  }
  return s
}

const randClock = () =>
  `${fmtTime(c.hour({ twentyfour: true }))}:${fmtTime(c.minute())}`

function deleteSchedule(id: string): Cypress.Chainable<void> {
  const query = `
    mutation deleteSchedule($input: DeleteScheduleInput!){
      deleteSchedule(input: $input) { deleted_id }
    }
  `

  return cy.graphql(query, { input: { id } })
}

Cypress.Commands.add('createSchedule', createSchedule)
Cypress.Commands.add('setScheduleTarget', setScheduleTarget)
Cypress.Commands.add('deleteSchedule', deleteSchedule)
