import { testScreen } from '../support'
import { DateTime } from 'luxon'

testScreen('Calendar', testCalendar)

const monthHeaderFormat = (t: DateTime) => t.toFormat('MMMM')
const weekHeaderFormat = (t: DateTime) => {
  const start = t.startOf('week').minus({ day: 1 })

  const end = t.endOf('week').minus({ day: 1 })

  return (
    start.toFormat('MMMM dd - ') +
    end.toFormat(end.month === start.month ? 'dd' : 'MMMM dd')
  )
}

function testCalendar(screen: ScreenFormat) {
  if (screen !== 'widescreen') return

  let sched: Schedule
  let rot: Rotation

  let now: DateTime
  beforeEach(() => {
    now = DateTime.local()
    cy.createSchedule().then(s => {
      sched = s

      cy.createRotation({
        count: 3,
        type: 'hourly',
        shiftLength: 12,
      }).then(r => {
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
        }).then(() => cy.visit('/schedules/' + sched.id))
      })
    })
  })

  it('should view shifts', () => {
    for (let i = 0; i < rot.users.length; i++) {
      cy.get('body').should('contain', rot.users[i].name.split(' ')[0])
    }
  })

  it(`should view a shift's tooltip`, () => {
    cy.get('div')
      .contains(rot.users[0].name.split(' ')[0])
      .trigger('mouseover')
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

  it.skip('should switch between weekly and monthly views', () => {
    cy.get('button[data-cy="show-month"]').should('be.disabled')
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      monthHeaderFormat(now),
    )

    cy.get('button[data-cy="show-week"]').click()
    cy.get('button[data-cy="show-week"]').should('be.disabled')
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      weekHeaderFormat(now),
    )

    cy.get('button[data-cy="show-month"]').click()
    cy.get('button[data-cy="show-month"]').should('be.disabled')
    cy.get('[data-cy="calendar-header"]').should(
      'contain',
      monthHeaderFormat(now),
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
    cy.fixture('users').then(users => {
      cy.get('button[data-cy="add-override"]').click()
      cy.get('input[name=addUserID]').selectByLabel(users[0].name)
      cy.get('button[type="submit"]').click()
      cy.get('button[type="submit"]').should('not.exist')
    })
  })

  it('should create a replace override from a shift tooltip', () => {
    cy.fixture('users').then(users => {
      cy.get('button[data-cy="add-override"]').click()
      cy.get('input[name=addUserID]').selectByLabel(users[0].name)
      cy.get('button[type="submit"]').click()
      cy.get('button[type="submit"]').should('not.exist')
    })
  })

  it('should create a remove override from a shift tooltip', () => {
    cy.get('div')
      .contains(rot.users[0].name.split(' ')[0])
      .trigger('mouseover')
    cy.get('div[data-cy="shift-tooltip"]').should('be.visible')
    cy.get('button[data-cy="remove-override"]').click()
    cy.get('button[type="submit"]').click()
    cy.get('button[type="submit"]').should('not.exist')
  })
}
