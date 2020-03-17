import { Chance } from 'chance'
import { DateTime } from 'luxon'

import { testScreen } from '../support'
const c = new Chance()

testScreen('Time Pickers', testTimePickers)

function testTimePickers() {
  describe('Time (schedule assignments)', () => {
    const check = (name: string, params: string, display: string) =>
      it(name, () => {
        cy.setScheduleTarget({
          schedule: { timeZone: 'America/New_York' },
          rules: [
            {
              weekdayFilter: [true, false, false, false, false, false, false],
              start: '15:04',
              end: '04:23',
            },
          ],
        }).then(tgt => {
          return cy.visit(`/schedules/${tgt.schedule.id}/assignments${params}`)
        })
        // sanity check
        cy.get('body').contains(display)

        cy.get('button[data-cy=other-actions]').menu('Edit')

        cy.dialogTitle('Edit Rules for Rotation')

        // check display
        cy.dialogForm({
          'rules[0].start': '14:05',
          'rules[0].end': '17:56',
        })

        cy.dialogFinish('Submit')

        cy.get('body').contains('Sun from 2:05 PM to 5:56 PM')
      })

    describe('Native', () => {
      check(
        'should handle selecting time values when displaying the same time zone',
        '?tz=America/New_York',
        'Sun from 3:04 PM to 4:23 AM',
      )
      check(
        'should handle selecting time values when displaying an alternate time zones',
        '?tz=America/Boise',
        'Sun from 1:04 PM to 2:23 AM',
      )
    })

    describe('Fallback', () => {
      check(
        'should handle selecting time values when displaying the same time zone',
        '?tz=America/New_York&nativeInput=0',
        'Sun from 3:04 PM to 4:23 AM',
      )
      check(
        'should handle selecting time values when displaying an alternate time zones',
        '?tz=America/Boise&nativeInput=0',
        'Sun from 1:04 PM to 2:23 AM',
      )
    })
  })

  describe('Date (schedule shifts)', () => {
    const check = (name: string, params: string, display: string) =>
      it(name, () => {
        cy.createSchedule({ timeZone: 'America/New_York' }).then(s =>
          cy.visit(`/schedules/${s.id}/shifts${params}`),
        )

        // sanity check
        cy.get('body').contains(display)

        cy.get('button[title=Filter]').click()
        cy.form({ filterStart: '2007-02-03' })

        cy.get('body').contains('2/3/2007')
      })

    describe('Native', () => {
      check(
        'should handle selecting date values when displaying the same time zone',
        '?tz=America/New_York&start=2006-01-02T06%3A00%3A00.000Z',
        '1/2/2006',
      )
      check(
        'should handle selecting date values when displaying an alternate time zone',
        '?tz=America/Boise&start=2006-01-02T06%3A00%3A00.000Z',
        '1/1/2006',
      )
    })

    describe('Fallback', () => {
      check(
        'should handle selecting date values when displaying the same time zone',
        '?tz=America/New_York&start=2006-01-02T06%3A00%3A00.000Z&nativeInput=0',
        '1/2/2006',
      )
      check(
        'should handle selecting date values when displaying an alternate time zone',
        '?tz=America/Boise&start=2006-01-02T06%3A00%3A00.000Z&nativeInput=0',
        '1/1/2006',
      )
    })
  })

  describe('DateTime (schedule overrides)', () => {
    const check = (name: string, params: string) =>
      it(name, () => {
        cy.createSchedule({ timeZone: 'America/New_York' }).then(s =>
          cy.visit(`/schedules/${s.id}/overrides${params}`),
        )

        cy.pageFab('Add a User')
        cy.dialogTitle('Temporarily Add a User')
        const month = DateTime.utc().month
        const start = DateTime.fromJSDate(
          c.date({
            year: DateTime.utc().year + 2,
            month: month === 12 ? 11 : month + 1,
          }) as Date,
        )

        cy.fixture('users').then(users => {
          const userName: string = (c.pickone(users) as Profile).name
          cy.dialogForm({
            addUserID: userName,
            start,
            end: start.plus({ days: 1 }),
          })
        })

        cy.dialogFinish('Submit')

        // sanity check
        cy.get('body').contains(start.toLocaleString(DateTime.DATETIME_MED))
      })

    describe('Native', () => {
      check(
        'should handle selecting date values when displaying the same time zone',
        '?tz=America/New_York',
      )
      check(
        'should handle selecting date values when displaying an alternate time zone',
        '?tz=America/Boise',
      )
    })

    describe('Fallback', () => {
      check(
        'should handle selecting date values when displaying the same time zone',
        '?tz=America/New_York&start=2006-01-02T06%3A00%3A00.000Z&nativeInput=0',
      )
      check(
        'should handle selecting date values when displaying an alternate time zone',
        '?tz=America/Boise&start=2006-01-02T06%3A00%3A00.000Z&nativeInput=0',
      )
    })
  })
}
