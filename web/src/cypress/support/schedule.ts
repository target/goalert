import { Chance } from 'chance'
import { DateTime, Interval } from 'luxon'
import {
  Schedule,
  ScheduleTarget,
  ScheduleTargetInput,
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
): Cypress.Chainable<void> {
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

  const input = options || {}
  const scheduleDuration = c.integer({ min: 1, max: 30 })
  const cur = DateTime.local()
  const curDay = cur.day
  const curMonth = cur.month
  const curYear = cur.year

  if (!input.start && !input.end) {
    // set start to anytime between now and 3 years (arbitrary but not too far in the future)
    input.start = DateTime.fromJSDate(c.date({
      year: curYear + c.integer({ min: 0, max: 3 })
    }) as Date)
  } else if (!input.start && input.end) {
    // set start to a random duration before end, if end is set
    input.start = DateTime.fromISO(input.end).minus({ days: scheduleDuration }).toISO()
  }

  if (!input.end) {
    // set end to a random duration after start
    const start = DateTime.fromISO(input.start as string)
    input.end = start.plus({ days: scheduleDuration }).toISO()
  }

  if (!input.shifts?.length) {
    input.shifts = []
    // addShifts adds a shift for each user specified to the input
    const addShifts = (users: string[]) => {
      const s = DateTime.fromISO(input.start)
      const e = DateTime.fromISO(input.end)

      for(let i = 0; i < users.length; i++) {
        const startYear = c.integer({ min: s.year, max: e.year })
        const startMonth = c.integer({ min: curYear === startYear ? s.month : 1, max: curYear === e.year ? e.month : 12 })
        const start = DateTime.fromObject({
          year: startYear,
          month: startMonth,
          day: c.integer({ min: curYear === startYear && curMonth === startMonth ? curDay : 0, max: curMonth === e.month ? e.day : DateTime.local(startYear, startMonth).daysInMonth }),
        })

        console.log('max: ', Interval.fromDateTimes(start, e).toDuration('hours'))
        
        input.shifts.push({
          start: start.toISO(), // anytime between (input.start and input.end) - scheduleDuration
          end: start.plus({ hours: c.floating({ min: 0.25, max: Interval.fromDateTimes(start, e).toDuration('hours').hours }) }), // anytime after set start and before input.end, random duration
          userID: users[i]
        })
      }
    }

    if (input.shiftUserIDs.length) {
      addShifts(input.shiftUserIDs)
    } else {
      cy.fixture('users').then((_users) => {
        const numUsers = c.integer({ min: 1, max: _users.length })
        let users = _users.slice()
        users.splice(numUsers - 1, 1)
        addShifts(users)
      })
    }
  }

  return cy.graphql(mutation, { input })
}

Cypress.Commands.add('createSchedule', createSchedule)
Cypress.Commands.add('setScheduleTarget', setScheduleTarget)
Cypress.Commands.add('deleteSchedule', deleteSchedule)
Cypress.Commands.add('createTemporarySchedule', createTemporarySchedule)
