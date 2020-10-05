import { testScreen } from '../../support'
import { DateTime } from 'luxon'
import { Schedule } from '../../../schema'

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
        count: 3,
        type: 'hourly',
        shiftLength: 1,
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

  it('should view shifts', () => {
    let check = rot.users.length
    // TODO: This could still fail between 10pm and 11:59pm
    // on the last day of the month (since the next day/shift isn't rendered)
    //
    // Once the calendar render fixes are in, it could still happen if the last day
    // of the month is a Saturday.
    //
    // Until then, it will also fail on the last day of any month based on the current time.
    //
    // Proper fix would be to control the time (frontend and backend) when these tests are run
    // to explicitly (and predictably) check these edge cases.

    if (now.endOf('month').day === now.day) {
      if (now.hour >= 11) {
        check = 1
      } else if (now.hour >= 10) {
        check = 2
      }
    }

    for (let i = 0; i < check; i++) {
      cy.get('body').should('contain', rot.users[i].name)
    }
  })

  it(`should view a shift's tooltip`, () => {
    cy.get('div').contains(rot.users[0].name).trigger('mouseover')
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

  it('should add an override from the calendar', () => {
    cy.fixture('users').then((users) => {
      cy.get('button[data-cy="add-override"]').click()
      cy.dialogTitle('Add a User')
      cy.dialogForm({ addUserID: users[0].name })
      cy.dialogFinish('Submit')
    })
  })

  it('should create a replace override from a shift tooltip', () => {
    const name = rot.users[0].name

    cy.fixture('users').then((users) => {
      let addUserName = users[0].name
      if (rot.users[0].id === users[0].id) addUserName = users[1].name
      cy.get('[data-cy=calendar]')
        .should('contain', name)
        .contains('div', name)
        .trigger('mouseover')
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
      .trigger('mouseover')
    cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
    cy.get('button[data-cy="remove-override"]').click()
    cy.dialogTitle('Remove a User')
    cy.dialogFinish('Submit')
  })
}

testScreen('Calendar', testCalendar)
