import { Chance } from 'chance'
import { testScreen } from '../../support'
import { Schedule, User } from '../../../schema'
import { DateTime, Interval } from 'luxon'
import { round } from 'lodash-es'

const c = new Chance()
const dtFmt = "yyyy-MM-dd'T'HH:mm"

function makeIntervalDates(): [string, string, number] {
  const now = DateTime.local()
  // year is either between now and 3 years in the future
  const year = c.integer({
    min: now.year,
    max: now.year + c.integer({ min: 1, max: 3 }),
  })
  // random month, keep in the future if using current year
  const month = c.integer({
    min: year === now.year ? now.month : 1,
    max: 12,
  })
  // random day, keep in future if using current month
  const day = c.integer({
    min: month === now.month ? now.day : 1,
    max: now.daysInMonth,
  })

  // create start, create end time from start
  const start = DateTime.fromObject({ year, month, day }).startOf('day')
  const end = start
    .plus({
      days: c.integer({ min: 1, max: 31 }),
    })
    .endOf('day')

  const duration = Interval.fromDateTimes(
    start,
    end.minus({ minute: 1 }),
  ).toDuration('hours')

  return [start.toFormat(dtFmt), end.toFormat(dtFmt), round(duration.hours, 2)]
}

function testTemporarySchedule(screen: ScreenFormat): void {
  let schedule: Schedule
  let manualAddUser: User
  let graphQLAddUser: User
  let schedAssignmentUser: User
  beforeEach(() => {
    cy.fixture('users').then((u) => {
      manualAddUser = u[0]
      graphQLAddUser = u[1]
      schedAssignmentUser = u[2]

      cy.createSchedule().then((s: Schedule) => {
        schedule = s
        cy.visit('/schedules/' + s.id)
      })
    })
  })

  it('should go back and forth between steps', () => {
    const [start, end] = makeIntervalDates()
    cy.get('[data-cy="new-temp-sched"]').click()
    cy.get('div[data-cy="sched-times-step"]').should('be.visible')
    cy.get('[data-cy="loading-button"]').contains('Next').click() // should error
    cy.focused().blur() // dismiss error
    cy.dialogForm({ start, end }, 'div[data-cy="sched-times-step"]')
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get('div[data-cy="add-shifts-step"]').should('be.visible')
    cy.dialogClick('Back')
    cy.get('div[data-cy="sched-times-step"]').should('be.visible')
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get('div[data-cy="add-shifts-step"]').should('be.visible')
  })

  const datePlusEight = (dt: string) => DateTime.fromFormat(dt, dtFmt).plus({ hours: 8 }).toFormat(dtFmt)

  it('should toggle duration field', () => {
    const [start, end] = makeIntervalDates()
    const shiftEnd = datePlusEight(start)
    cy.get('[data-cy="new-temp-sched"]').click()
    cy.dialogForm({ start, end }, 'div[data-cy="sched-times-step"]')
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get('div[data-cy="add-shifts-step"]').should('be.visible')
    cy.get('div[data-cy="add-shifts-step"] input[name="end"]').should('have.value', 8)
    cy.get('div[data-cy="add-shifts-step"] span[data-cy="toggle-duration-off"]').click()
    cy.get('div[data-cy="add-shifts-step"] input[name="end"]').should('have.value', shiftEnd)
    cy.dialogForm({ end: datePlusEight(shiftEnd) }, 'div[data-cy="add-shifts-step"]')
    cy.get('div[data-cy="add-shifts-step"] span[data-cy="toggle-duration-on"]').click()
    cy.get('div[data-cy="add-shifts-step"] input[name="end"]').should('have.value', 16)
  })

  it('should toggle timezone switches', () => {
    const c = (t: string, tz: string) => {
      let dt = DateTime.fromFormat(t, dtFmt)
      dt = dt.setZone(tz)
      console.log(dt)
      return dt.toFormat(dtFmt)
    }
    const lTZ = (t: string): string => c(t, DateTime.local().zoneName)
    const sTZ = (t: string): string => c(t, schedule.timeZone)
    
    const [start, end] = makeIntervalDates()

    cy.get('[data-cy="new-temp-sched"]').click()
    cy.dialogForm({ start, end }, 'div[data-cy="sched-times-step"]')
    cy.get('div[data-cy="sched-times-step"] input[name="start"]').should('have.value', lTZ(start))
    cy.get('div[data-cy="sched-times-step"] input[name="end"]').should('have.value', lTZ(end))
    cy.get('div[data-cy="sched-times-step"] [data-cy="tz-switch"]').click()
    cy.get('div[data-cy="sched-times-step"] input[name="start"]').should('have.value', sTZ(start))
    cy.get('div[data-cy="sched-times-step"] input[name="end"]').should('have.value', sTZ(end))
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get('div[data-cy="add-shifts-step"]').should('be.visible')
    cy.get('div[data-cy="add-shifts-step"] [data-cy="toggle-duration-off"]').click()
    cy.get('div[data-cy="add-shifts-step"] input[name="start"]').should('have.value', sTZ(start))
    cy.get('div[data-cy="add-shifts-step"] input[name="end"]').should('have.value', sTZ(datePlusEight(start)))
    cy.get('div[data-cy="add-shifts-step"] [data-cy="tz-switch"]').click()
    cy.get('div[data-cy="add-shifts-step"] input[name="start"]').should('have.value', lTZ(start))
    cy.get('div[data-cy="add-shifts-step"] input[name="end"]').should('have.value', lTZ(datePlusEight(start)))
    cy.dialogClick('Back')
    cy.get('div[data-cy="sched-times-step"] input[name="start"]').should('have.value', lTZ(start))
    cy.get('div[data-cy="sched-times-step"] input[name="end"]').should('have.value', lTZ(end))
  })

  it('should refill a shifts info after deleting in step 2', () => {
    cy.createTemporarySchedule(schedule.id).then(() => {
      cy.reload()
      cy.get('div').contains('Temporary Schedule').trigger('mouseover')
      cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
      cy.get('button[data-cy="edit-temp-sched"]').click()
      cy.get('[data-cy="shifts-list"]').should('contain', graphQLAddUser.name)
      cy.get('div[data-cy="add-shifts-step"] input[name="userID"]').should('have.value', '')
      // todo: check start and end input values before deleting
      cy.get('[data-cy="shifts-list"] li').contains(graphQLAddUser.name).eq(0).parent().parent().siblings().click() // delete
      cy.get('div[data-cy="add-shifts-step"] input[name="userID"]').should('have.value', graphQLAddUser.name)
      // todo: check start and end input values from data
    })
  })

  it('should cancel and close form', () => {
    cy.get('[role="dialog"]').should('not.exist')
    cy.get('[data-cy="new-temp-sched"]').click()
    cy.get('[role="dialog"]').should('be.visible')
    cy.dialogFinish('Cancel')
    cy.get('[role="dialog"]').should('not.exist')
  })

  // todo: start with shifts on schedule and check they disappear after creating
  it('should create a temporary schedule', () => {
    // note: could check calendar for original shift in weekly view
    // it would us to compare shift times with a user's name in the same div without having to open a tooltip
    const [start, end, duration] = makeIntervalDates()
    cy.get('[data-cy="new-temp-sched"]').click()
    cy.dialogForm({ start, end }, 'div[data-cy="sched-times-step"]')
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get('div[data-cy="add-shifts-step"]').should('be.visible')
    cy.dialogForm(
      {
        start,
        end: duration,
        userID: manualAddUser.name,
      },
      'div[data-cy="add-shifts-step"]',
    )
    cy.get('[data-cy="shifts-list"]').should('not.contain', manualAddUser.name)
    cy.get('button[title="Add Shift"]').click()
    cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)
    cy.dialogFinish('Submit')
    cy.visit(
      '/schedules/' +
        schedule.id +
        '?start=' +
        DateTime.fromISO(start).toFormat('yyyy-MM-dd') +
        'T07%3A00%3A00.000Z',
    )
    cy.get('div').contains('Temporary Schedule').trigger('mouseover')
    cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
    cy.get('button[data-cy="edit-temp-sched"]').should('be.visible')
    cy.get('button[data-cy="delete-temp-sched"]').should('be.visible')
    cy.get('div').contains(manualAddUser.name).trigger('mouseover')
    cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
    // todo
    // find by day then name to verify as temporary? (eq[0] since business logic = should always be sorted as first in calendar)
    // check by color being green
  })

  it('should edit a temporary schedule', () => {
    cy.createTemporarySchedule(schedule.id).then(() => {
      cy.reload()
      cy.get('div').contains('Temporary Schedule').trigger('mouseover')
      cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
      cy.get('button[data-cy="edit-temp-sched"]').click()
      cy.get('[data-cy="shifts-list"]').should('contain', graphQLAddUser.name)
      cy.get('[data-cy="shifts-list"] li').contains(graphQLAddUser.name).eq(0).parent().parent().siblings().click() // delete
      cy.get('[data-cy="shifts-list"]').should('not.contain', graphQLAddUser.name)
      cy.dialogForm({ userID: manualAddUser.name }, 'div[data-cy="add-shifts-step"]')
      cy.get('[data-cy="shifts-list"]').should('not.contain', manualAddUser.name)
      cy.get('button[title="Add Shift"]').click()
      cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)
      cy.dialogFinish('Submit')
      cy.reload() // ensure calendar update
      cy.get('div').contains(manualAddUser.name).trigger('mouseover')
      cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
      // todo
      // find by day then name to verify as temporary? (eq[0] since business logic = should always be sorted as first in calendar)
      // check by color being green
    })
  })

  it('should delete a temporary schedule', () => {
      cy.createTemporarySchedule(schedule.id).then(() => {
        cy.reload()
        // todo: check original schedule assignment shifts in calendar
        cy.get('div').contains('Temporary Schedule').trigger('mouseover')
        cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
        cy.get('button[data-cy="delete-temp-sched"]').click()
        cy.dialogFinish('Confirm')
        cy.get('div').contains('Temporary Schedule').should('not.exist')
        // todo: check that original shifts show again
      })
  })

  it('should be able to add multiple shifts on step 2', () => {
    const [start, end, duration] = makeIntervalDates()
    cy.get('[data-cy="new-temp-sched"]').click()
    cy.dialogForm({ start, end }, 'div[data-cy="sched-times-step"]')
    cy.get('[data-cy="loading-button"]').contains('Next').click()
    cy.get('div[data-cy="add-shifts-step"]').should('be.visible')
    cy.dialogForm(
      {
        start,
        end: duration / 2,
        userID: manualAddUser.name,
      },
      'div[data-cy="add-shifts-step"]',
    )
    cy.get('[data-cy="shifts-list"]').should('not.contain', manualAddUser.name)
    cy.get('button[title="Add Shift"]').click()
    cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)
    cy.dialogForm(
      {
        userID: graphQLAddUser.name, // gql not in this test, safe to use here
      },
      'div[data-cy="add-shifts-step"]',
    )
    cy.get('[data-cy="shifts-list"]').should('not.contain', graphQLAddUser.name)
    cy.get('button[title="Add Shift"]').click()
    cy.get('[data-cy="shifts-list"]').should('contain', graphQLAddUser.name)
    cy.get('[data-cy="shifts-list"]').should('contain', manualAddUser.name)
    // todo: verify list is sorted
  })
}

testScreen('temporary Schedule', testTemporarySchedule)
