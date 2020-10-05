import { Chance } from 'chance'
import { DateTime } from 'luxon'
import {
  Schedule,
  ScheduleTarget,
  ScheduleTargetInput,
  FixedShiftGroup,
} from '../../schema'

const c = new Chance()

const fmtTime = (num: number): string => {
  const s = num.toString()
  if (s.length === 1) {
    return '0' + s
  }
  return s
}

const randClock = (): string =>
  `${fmtTime(c.hour({ twentyfour: true }))}:${fmtTime(c.minute())}`

function setScheduleTarget(
  scheduleTgt?: Partial<ScheduleTargetInput>,
  createScheduleInput?: Partial<Schedule>,
): Cypress.Chainable<ScheduleTarget> {
  if (!scheduleTgt) {
    scheduleTgt = {}
  }
  if (!scheduleTgt.scheduleID) {
    return cy
      .createSchedule(createScheduleInput)
      .then((sched: Schedule) =>
        setScheduleTarget({ ...scheduleTgt, scheduleID: sched.id }),
      )
  }
  if (!scheduleTgt.target) {
    return cy.createRotation().then((r: Rotation) =>
      setScheduleTarget({
        ...scheduleTgt,
        target: { type: 'rotation', id: r.id },
      }),
    )
  }
  if (!scheduleTgt.rules) {
    scheduleTgt.rules = [{}]
  }
  scheduleTgt.rules = scheduleTgt.rules.map((r) => ({
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
      isFavorite
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

  const params = scheduleTgt

  return cy
    .graphql(mutation, {
      input: params,
    })
    .then(() => {
      return cy
        .graphql(query, {
          id: params.scheduleID,
          tgt: params.target,
        })
        .then((res: GraphQLResponse) => {
          const { target, ...schedule } = res.schedule
          return {
            ...target,
            scheduleID: schedule.id,
          }
        })
    })
}

function createSchedule(
  sched?: Partial<Schedule>,
): Cypress.Chainable<Schedule> {
  const query = `mutation createSchedule($input: CreateScheduleInput!){
      createSchedule(input: $input) {
        id
        name
        description
        timeZone
        isFavorite
      }
    }`

  if (!sched) sched = {}

  return cy
    .graphql(query, {
      input: {
        name: sched.name || 'SM Sched ' + c.word({ length: 8 }),
        description: sched.description || c.sentence(),
        timeZone: sched.timeZone || 'America/Chicago',
        favorite: sched.isFavorite,
      },
    })
    .then((res: GraphQLResponse) => res.createSchedule)
}

function deleteSchedule(id: string): Cypress.Chainable<void> {
  const mutation = `
    mutation($input: [TargetInput!]!) {
      deleteAll(input: $input)
    }
  `

  return cy.graphql(mutation, {
    input: [
      {
        type: 'schedule',
        id,
      },
    ],
  })
}

function createFixedSchedule(
  scheduleID?: string,
  options?: Partial<FixedShiftGroup>,
): Cypress.Chainable<void> {
  const mutation = `
    mutation($input: SetScheduleShiftsInput!) {
      setScheduleShifts(input: $input)
    }
  `

  if (!scheduleID) {
    cy.createSchedule().then((s: Schedule) =>
      createFixedSchedule(s.id, options),
    )
  }

  const nowDT = DateTime.local()
  let input = options || {}
  if (!input.start) {
    input.start = nowDT.startOf('day').toISO()
  }
  if (!input.end) {
    input.end = nowDT.plus({ days: 7 }).endOf('day').toISO()
  }
  if (!input.shifts?.length) {
    cy.fixture('users').then((users) => {
      input.shifts = [
        {
          start: '',
          end: '',
          userID: users[0].id,
          truncated: false,
        },
      ]
    })
  }

  return cy.graphql(mutation, { input })
}

Cypress.Commands.add('createSchedule', createSchedule)
Cypress.Commands.add('setScheduleTarget', setScheduleTarget)
Cypress.Commands.add('deleteSchedule', deleteSchedule)
