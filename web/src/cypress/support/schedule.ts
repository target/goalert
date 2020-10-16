import { Chance } from 'chance'
import { DateTime, Interval } from 'luxon'
import { omit } from 'lodash-es'
import {
  Schedule,
  ScheduleTarget,
  ScheduleTargetInput,
  User,
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

function createTemporarySchedule(
  scheduleID?: string,
  options?: TemporaryScheduleOptions,
): Cypress.Chainable<null> {
  const mutation = `
    mutation($input: SetTemporaryScheduleInput!) {
      setTemporarySchedule(input: $input)
    }
  `

  if (!scheduleID) {
    return cy
      .createSchedule()
      .then((s: Schedule) => createTemporarySchedule(s.id, options))
  }

  let input = omit(options ?? {}, 'shiftUserIDs')
  input.scheduleID = scheduleID

  const now = DateTime.local()
  const r = (min: number, max: number): number => c.integer({ min, max })

  const MAX_FUTURE = 1576800 // up to 3 years (in minutes) in the future
  const MIN = 60 // minimum temp sched length of 1 hour, in minutes
  const MAX = 43800 // maximum temp sched length of 1 month, in minutes
  const S_MIN = 1 // minimum shift length, in hours

  // set temp sched start and end dates
  if (!input.start && !input.end) {
    let s = now.plus({ minutes: r(0, MAX_FUTURE) })
    input.start = s.toISO()
    input.end = s.plus({ minutes: r(MIN, MAX) }).toISO()
  } else if (!input.start && input.end) {
    const end = DateTime.fromISO(input.end)
    if (!end.isValid) return cy.log('invalid end date')
    if (+end < +now) return cy.log('cannot provide end time before now() without also providing start time')
    const max = Interval.fromDateTimes(now, end).toDuration('hours').hours
    input.start = end.minus({ hours: r(1, max) }).toISO()
  } else if (input.start && !input.end) {
    const start = DateTime.fromISO(input.start)
    if (!start.isValid) return cy.log('invalid start date')
    input.end = start.plus({ minutes: r(MIN, MAX) }).toISO()
  }

  // set shifts
  if (!input.shifts?.length) {
    cy.fixture('users').then((users) => {
      const userIDs = options.shiftUserIDs || c.pickset(users.map((u: User) => u.id), r(1, users.length))
      const schedStart = DateTime.fromISO(input.start)
      const schedEnd = DateTime.fromISO(input.end)
      if (!schedStart.isValid) return cy.log('invalid start date')
      if (!schedEnd.isValid) return cy.log('invalid end date')

      if (+schedStart > +schedEnd) return cy.log('start cannot begin after end')
      const schedLength = Interval.fromDateTimes(schedStart, schedEnd).toDuration(['hours', 'minutes'])

      // make 1 shift per user, within range of sched
      input.shifts = []
      userIDs.forEach((userID: string) => {

        const start = schedStart.plus({ minutes: r(0, schedLength.minutes - S_MIN) })
        const timeUntilEnd = Interval.fromDateTimes(start, schedEnd).toDuration('minutes').minutes
        const end = start.plus({ minutes: r(S_MIN, timeUntilEnd) })

        input.shifts.push({ userID, start, end })
      })
    })
  }

  return cy.graphql(mutation, { input })
}

Cypress.Commands.add('createSchedule', createSchedule)
Cypress.Commands.add('setScheduleTarget', setScheduleTarget)
Cypress.Commands.add('deleteSchedule', deleteSchedule)
Cypress.Commands.add('createTemporarySchedule', createTemporarySchedule)
