import { Chance } from 'chance'
import { DateTime, Interval } from 'luxon'
import {
  Schedule,
  ScheduleTarget,
  ScheduleTargetInput,
  SetScheduleShiftInput,
  SetTemporaryScheduleInput,
  User,
} from '../../schema'
import { randDT, randSubInterval } from './util'

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

function genShifts(
  userIDs: string[],
  start: DateTime,
  end: DateTime,
  _shifts?: Partial<SetScheduleShiftInput>[],
): SetScheduleShiftInput[] {
  const shifts = _shifts || new Array(c.integer({ min: 0, max: 10 })).fill({})
  if (shifts.length === 0) return []

  const schedIvl = Interval.fromDateTimes(start, end)
  return schedIvl.divideEqually(shifts.length).map((ivl, i) => {
    const shift = shifts[i]
    const rIvl = randSubInterval(ivl)
    return {
      userID: shift.userID || c.pickone(userIDs),
      start: shift.start || rIvl.start.toISO(),
      end: shift.end || rIvl.end.toISO(),
    }
  })
}

function shiftRange(
  shifts?: Partial<SetScheduleShiftInput>[],
): [DateTime, DateTime] {
  if (!shifts || !shifts.length)
    return [DateTime.fromISO(''), DateTime.fromISO('')]

  let min = DateTime.fromISO('')
  let max = DateTime.fromISO('')
  shifts.forEach((s) => {
    const start = DateTime.fromISO(s.start || '')
    const end = DateTime.fromISO(s.end || '')

    if ((start.isValid && !min.isValid) || start < min) {
      min = start
    }

    if ((end.isValid && !max.isValid) || end > max) {
      max = end
    }
  })

  return [min, max]
}

interface SetTemporarySchedule {
  scheduleID: string
  schedule: Partial<Schedule>
  start: string
  end: string
  shifts: Partial<SetScheduleShiftInput>[]
}

function createTemporarySchedule(
  opts: Partial<SetTemporarySchedule> = {},
): Cypress.Chainable<null> {
  const mutation = `
    mutation($input: SetTemporaryScheduleInput!) {
      setTemporarySchedule(input: $input)
    }
  `

  // create schedule if necessary
  if (!opts.scheduleID) {
    return cy
      .createSchedule(opts.schedule)
      .then((s: Schedule) =>
        createTemporarySchedule({ ...opts, scheduleID: s.id }),
      )
  }

  const [shiftStart, shiftEnd] = shiftRange(opts.shifts)

  // set start/end time if necessary
  const now = DateTime.utc().plus({ hour: 1 })
  let start = DateTime.fromISO(opts.start || '')
  let end = DateTime.fromISO(opts.end || '')
  if (start.isValid && !end.isValid) {
    if (shiftEnd.isValid) {
      end = randDT({ min: shiftEnd })
    } else {
      end = randDT({ min: start.plus({ day: 1 }) })
    }
  } else if (end.isValid && !start.isValid) {
    if (shiftStart.isValid) {
      start = randDT({ min: now, max: shiftStart })
    } else {
      start = randDT({
        min: now,
        max: end.plus({ hour: -8 }),
      })
    }
  } else if (!start.isValid && !end.isValid) {
    start = now.plus({ days: c.floating({ min: 1, max: 3 }) })
    end = start.plus({ days: c.floating({ min: 2, max: 4 }) })
  }
  if (!start.isValid) throw new Error('invalid start time')
  if (!end.isValid) throw new Error('invalid end time')

  return cy.fixture('users').then((users) => {
    const userIDs: string[] = users.map((u: User) => u.id)
    const shifts = genShifts(userIDs, start, end, opts.shifts)

    const input: SetTemporaryScheduleInput = {
      scheduleID: opts.scheduleID as string, // checked above
      start: start.toISO(),
      end: end.toISO(),
      shifts,
    }

    return cy.graphql(mutation, { input })
  })
}

Cypress.Commands.add('createSchedule', createSchedule)
Cypress.Commands.add('setScheduleTarget', setScheduleTarget)
Cypress.Commands.add('deleteSchedule', deleteSchedule)
Cypress.Commands.add('createTemporarySchedule', createTemporarySchedule)
