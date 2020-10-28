import { Chance } from 'chance'
import { testScreen } from '../support'
import { Schedule, User } from '../../schema'
import { DateTime, Interval } from 'luxon'
import { round } from 'lodash-es'

const c = new Chance()
const dtFmt = "yyyy-MM-dd'T'HH:mm"
const schedTimesSelector = 'div[data-cy="sched-times-step"]'
const addShiftsSelector = 'div[data-cy="add-shifts-step"]'

// makeIntervalDates creates an interval, returning the start
// end, and duration (in hours)
function makeIntervalDates(): [DateTime, DateTime, number] {
  const MAX_FUTURE = 60 * 24 * 365 * 3 // up to 3 years (in minutes) in the future
  const MIN = 60 * 24 * 30 // minimum temp sched length of 1 hour, in minutes
  const MAX = 43800 // maximum temp sched length of 1 month, in minutes

  const now = DateTime.local()
  const r = (min: number, max: number): number => c.integer({ min, max })

  const start = now.plus({ minutes: r(0, MAX_FUTURE) })
  const end = start.plus({ minutes: r(MIN, MAX) })

  const duration = Interval.fromDateTimes(start, end).toDuration('hours')

  return [start, end, round(duration.hours, 2)]
}

function testTemporarySchedule(): void {
  let schedule: Schedule
  let manualAddUser: User
  let graphQLAddUser: User
  beforeEach(() => {
    cy.fixture('users').then((u) => {
      manualAddUser = u[0]
      graphQLAddUser = u[1]

      cy.createSchedule().then((s: Schedule) => {
        schedule = s
        cy.visit('/schedules/' + s.id)
      })
    })
  })

  it('should go back and forth between steps', () => {
    const [start, end] = makeIntervalDates()
    cy.get('[data-cy="new-temp-sched"]').click()
    cy.get(schedTimesSelector).as('step1')
    cy.get(addShiftsSelector).as('step2')
    cy.get('@step1').should('be.visible.and.contain', 'STEP 1 OF 2')
    cy.get('[data-cy="loading-button"]').contains('Next').click() // should error
    cy.dialogForm({ start, end }, schedTimesSelector)
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get('@step2').should('be.visible.and.contain', 'STEP 2 OF 2')
    cy.dialogClick('Back')
    cy.get('@step1').should('be.visible')
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get('@step2').should('be.visible')
  })

  it('should toggle duration field', () => {
    const [start, end] = makeIntervalDates()
    const shiftEnd = start.plus({ hours: 8 })
    cy.get('[data-cy="new-temp-sched"]').click()
    cy.get(addShiftsSelector).as('step2')
    cy.dialogForm({ start, end }, schedTimesSelector)
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get('@step2').should('be.visible.and.contain', 'STEP 2 OF 2')
    cy.get('@step2').find('input[name="end"]').should('have.value', 8)
    cy.get('@step2').find('[data-cy="toggle-duration-off"]').click()
    cy.get('@step2')
      .find('input[name="end"]')
      .should('have.value', shiftEnd.toFormat(dtFmt))
    cy.dialogForm(
      { end: shiftEnd.plus({ hours: 8 }).toFormat(dtFmt) },
      addShiftsSelector,
    )
    cy.get('@step2').find('[data-cy="toggle-duration-on"]').click()
    cy.get('@step2').find('input[name="end"]').should('have.value', 16)
  })

  it('should toggle timezone switches', () => {
    const [start, end] = makeIntervalDates()
    const c = (t: DateTime, tz: string): string => t.setZone(tz).toFormat(dtFmt)
    const locTZ = (t: DateTime): string => c(t, DateTime.local().zoneName)
    const schedTZ = (t: DateTime): string => c(t, schedule.timeZone)

    cy.get('[data-cy="new-temp-sched"]').click()
    cy.get(schedTimesSelector).as('step1')
    cy.get(addShiftsSelector).as('step2')
    cy.dialogForm({ start, end }, schedTimesSelector)
    cy.get('@step1')
      .find('input[name="start"]')
      .should('have.value', locTZ(start))
    cy.get('@step1').find('input[name="end"]').should('have.value', locTZ(end))
    cy.get('@step1').find('[data-cy="tz-switch"]').click()
    cy.get('@step1')
      .find('input[name="start"]')
      .should('have.value', schedTZ(start))
    cy.get('@step1')
      .find('input[name="end"]')
      .should('have.value', schedTZ(end))
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get('@step2').should('be.visible.and.contain', 'STEP 2 OF 2')
    cy.get('@step2').find('[data-cy="toggle-duration-off"]').click()
    cy.get('@step2')
      .find('input[name="start"]')
      .should('have.value', schedTZ(start))
    cy.get('@step2')
      .find('input[name="end"]')
      .should('have.value', schedTZ(start.plus({ hours: 8 })))
    cy.get('@step2').find('[data-cy="tz-switch"]').click()
    cy.get('@step2')
      .find('input[name="start"]')
      .should('have.value', locTZ(start))
    cy.get('@step2')
      .find('input[name="end"]')
      .should('have.value', locTZ(start.plus({ hours: 8 })))
    cy.dialogClick('Back')
    cy.get('@step1')
      .find('input[name="start"]')
      .should('have.value', locTZ(start))
    cy.get('@step1').find('input[name="end"]').should('have.value', locTZ(end))
  })

  it("should add shift's info to form after deleting it from shift list", () => {
    cy.createTemporarySchedule(schedule.id, {
      start: DateTime.local().toISO(),
      shiftUserIDs: [graphQLAddUser.id],
    }).then(() => {
      cy.reload()
      cy.get('div').contains('Temporary Schedule').trigger('mouseover')
      cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
      cy.get('button[data-cy="edit-temp-sched"]').click()
      cy.get(addShiftsSelector).as('step2')
      cy.get('[data-cy="shifts-list"]').should('contain', graphQLAddUser.name)
      cy.get('@step2').find('input[name="userID"]').should('have.value', '')
      cy.get('[data-cy="shifts-list"] li [data-cy="delete-shift"]').click({
        force: true,
      }) // delete
      cy.get('@step2')
        .find('input[name="userID"]')
        .should('have.value', graphQLAddUser.name)
    })
  })

  it('should cancel and close form', () => {
    cy.get('[role="dialog"]').should('not.exist')
    cy.get('[data-cy="new-temp-sched"]').click()
    cy.get('[role="dialog"]').should('be.visible')
    cy.dialogFinish('Cancel')
    cy.get('[role="dialog"]').should('not.exist')
  })

  it('should create a temporary schedule', () => {
    const [start, end] = makeIntervalDates()
    cy.get('[data-cy="new-temp-sched"]').click()
    cy.dialogForm({ start, end }, schedTimesSelector)
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get(addShiftsSelector).should('be.visible.and.contain', 'STEP 2 OF 2')
    cy.dialogForm({ userID: manualAddUser.name }, addShiftsSelector)
    cy.get('[data-cy="shifts-list"]').should('not.contain', manualAddUser.name)
    cy.get('button[title="Add Shift"]').click()
    cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)
    cy.dialogFinish('Submit')
    cy.visit(
      '/schedules/' +
        schedule.id +
        '?start=' +
        start.toFormat('yyyy-MM-dd') +
        'T07%3A00%3A00.000Z',
    )
    cy.get('div').contains('Temporary Schedule').trigger('mouseover')
    cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
    cy.get('button[data-cy="edit-temp-sched"]').should('be.visible')
    cy.get('button[data-cy="delete-temp-sched"]').should('be.visible')
    cy.get('div').contains(manualAddUser.name).trigger('mouseover')
    cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
  })

  // seems buggy when shift list has overflow
  it('should edit a temporary schedule', () => {
    const now = DateTime.local()
    cy.createTemporarySchedule(schedule.id, {
      start: now.startOf('minute').toISO(),
      shiftUserIDs: [graphQLAddUser.id],
    }).then(() => {
      cy.reload()
      cy.get('div').contains('Temporary Schedule').trigger('mouseover')
      cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
      cy.get('button[data-cy="edit-temp-sched"]').click()
      cy.get('[data-cy="shifts-list"]').should('contain', graphQLAddUser.name)
      cy.get('[data-cy="shifts-list"] li [data-cy="delete-shift"]').click({
        force: true,
      }) // delete
      cy.get('[data-cy="shifts-list"]').should(
        'not.contain',
        graphQLAddUser.name,
      )
      cy.dialogForm(
        { userID: manualAddUser.name, start: now.toFormat(dtFmt) },
        addShiftsSelector,
      )
      cy.get('[data-cy="shifts-list"]').should(
        'not.contain',
        manualAddUser.name,
      )
      cy.get('button[title="Add Shift"]').click()
      cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)
      cy.dialogFinish('Submit')
      cy.reload() // ensure calendar update
      cy.get('div').contains(manualAddUser.name).trigger('mouseover')
      cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
    })
  })

  it('should delete a temporary schedule', () => {
    cy.createTemporarySchedule(schedule.id, {
      start: DateTime.local().toISO(),
      shiftUserIDs: [graphQLAddUser.id],
    }).then(() => {
      cy.reload()
      cy.get('div').contains('Temporary Schedule').trigger('mouseover')
      cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
      cy.get('button[data-cy="delete-temp-sched"]').click()
      cy.dialogFinish('Confirm')
      cy.get('div').contains('Temporary Schedule').should('not.exist')
    })
  })

  it('should be able to add multiple shifts on step 2', () => {
    const [start, end, duration] = makeIntervalDates()
    cy.get('[data-cy="new-temp-sched"]').click()
    cy.dialogForm({ start, end }, schedTimesSelector)
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get(addShiftsSelector).should('be.visible.and.contain', 'STEP 2 OF 2')
    cy.dialogForm(
      {
        userID: manualAddUser.name,
        start,
        end: duration / 2,
      },
      addShiftsSelector,
    )
    cy.get('[data-cy="shifts-list"]').should('not.contain', manualAddUser.name)
    cy.get('button[title="Add Shift"]').click()
    cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)
    cy.dialogForm(
      {
        userID: graphQLAddUser.name, // gql not in this test, safe to use here
      },
      addShiftsSelector,
    )
    cy.get('[data-cy="shifts-list"]').should('not.contain', graphQLAddUser.name)
    cy.get('button[title="Add Shift"]').click()
    cy.get('[data-cy="shifts-list"]').should('contain', graphQLAddUser.name)
    cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)
  })
}

testScreen('temporary Schedule', testTemporarySchedule)
