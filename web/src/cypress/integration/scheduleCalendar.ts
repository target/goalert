import { Chance } from 'chance'
import { testScreen } from '../support'
import { DateTime } from 'luxon'
import { Schedule } from '../../schema'

const c = new Chance()

const monthHeaderFormat = (t: DateTime): string => t.toFormat('MMMM')
const weekHeaderFormat = (t: DateTime): string => {
  const start = t.startOf('week').minus({ day: 1 })

  const end = t.endOf('week').minus({ day: 1 })

  return (
    start.toFormat('MMMM dd — ') +
    end.toFormat(end.month === start.month ? 'dd' : 'MMMM dd')
  )
}

const weekSpansTwoMonths = (t: DateTime): boolean => {
  const start = t.startOf('week').minus({ day: 1 })
  const end = t.endOf('week').minus({ day: 1 })
  return start.month !== end.month
}

function testCalendar(screen: ScreenFormat): void {
  if (screen !== 'widescreen') return

  let sched: Schedule
  let rot: Rotation

  let now: DateTime
  beforeEach(() => {
    now = DateTime.local()
    cy.createSchedule().then((s: Schedule) => {
      sched = s

      cy.createRotation({
        numUsers: 3,
        type: 'hourly',
        // based on production data the majority of rotations have a shiftLength of
        // 4-12 hours. Roughly 50% of these are 12 hours, so pick between 12 or 4-11 hours
        shiftLength: c.pickone([c.integer({ min: 4, max: 11 }), 12]),
      }).then((r: Rotation) => {
        rot = r

        cy.setScheduleTarget({
          scheduleID: s.id,
          target: {
            type: 'rotation',
            id: r.id,
          },
          rules: [
            {
              start: '12:00',
              end: '12:00',
              weekdayFilter: [true, true, true, true, true, true, true],
            },
          ],
        }).then(() => {
          cy.visit('/schedules/' + sched.id)
          cy.get('[data-cy=calendar]', { timeout: 30000 }).should('be.visible')
        })
      })
    })
  })

  it('should view shifts in month view', () => {
    cy.get('button[data-cy="next"]').click() // view shifts within the scope of a full month

    for (let i = 0; i < rot.users.length; i++) {
      cy.get('body [data-cy="calendar"]').should('contain', rot.users[i].name)
    }
  })

  it(`should view a shift's tooltip`, () => {
    cy.get('div').contains(rot.users[0].name).click()
    cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
    cy.get('button[data-cy="replace-override"]').should('be.visible')
    cy.get('button[data-cy="remove-override"]').should('be.visible')
  })

  it('should navigate by month', () => {
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      monthHeaderFormat(now),
    )
    cy.get('button[data-cy="next"]').click()
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      monthHeaderFormat(now.plus({ month: 1 })),
    )
    cy.get('button[data-cy="back"]').click()
    cy.get('button[data-cy="back"]').click()
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      monthHeaderFormat(now.minus({ month: 1 })),
    )
    cy.get('button[data-cy="show-today"]').click()
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      monthHeaderFormat(now),
    )
  })

  it('should switch between weekly and monthly views', () => {
    // defaults to current month
    cy.get('button[data-cy="show-month"]').should('be.disabled')
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      monthHeaderFormat(now),
    )

    // click weekly
    cy.get('button[data-cy="show-week"]').click()
    cy.get('button[data-cy="show-week"]').should('be.disabled')
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      weekHeaderFormat(now),
    )

    // go from week to monthly view
    // e.g. if navigating to an overlap of two months such as
    // Jan 27 - Feb 2, show the latter month (February)
    let monthsToAdd = 0
    if (weekSpansTwoMonths(now) && now.day > 7) {
      monthsToAdd = 1
    }

    cy.get('button[data-cy="show-month"]').click()
    cy.get('button[data-cy="show-month"]').should('be.disabled')
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      monthHeaderFormat(now.plus({ months: monthsToAdd })),
    )
  })

  it('should navigate by week', () => {
    cy.get('button[data-cy="show-week"]').click()
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      weekHeaderFormat(now),
    )
    cy.get('button[data-cy="next"]').click()
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      weekHeaderFormat(now.plus({ week: 1 })),
    )

    cy.get('button[data-cy="back"]').click()
    cy.get('button[data-cy="back"]').click()
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      weekHeaderFormat(now.minus({ week: 1 })),
    )
    cy.get('button[data-cy="show-today"]').click()
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      weekHeaderFormat(now),
    )
  })

  it('should create a replace override from a shift tooltip', () => {
    const name = rot.users[0].name

    cy.fixture('users').then((users) => {
      let addUserName = users[0].name
      if (rot.users[0].id === users[0].id) addUserName = users[1].name
      cy.get('[data-cy=calendar]')
        .should('contain', name)
        .contains('div', name)
        .click()
      cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
      cy.get('button[data-cy="replace-override"]').click()
      cy.dialogTitle('Replace a User')
      cy.dialogForm({ addUserID: addUserName })
      cy.dialogFinish('Submit')
    })
  })

  it('should create a remove override from a shift tooltip', () => {
    const name = rot.users[0].name

    cy.get('[data-cy=calendar]')
      .should('contain', name)
      .contains('div', name)
      .click()
    cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
    cy.get('button[data-cy="remove-override"]').click()
    cy.dialogTitle('Remove a User')
    cy.dialogFinish('Submit')
  })
}

testScreen('Calendar', testCalendar)
