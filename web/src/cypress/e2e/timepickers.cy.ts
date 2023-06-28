import { Chance } from 'chance'
import { DateTime } from 'luxon'
import { Schedule, ScheduleTarget } from '../../schema'
import users from '../fixtures/users.json'

import { testScreen } from '../support/e2e'
const c = new Chance()

function testTimePickers(screen: ScreenFormat): void {
  describe('Time (schedule assignments)', () => {
    const check = (name: string, params: string, display: string): Mocha.Test =>
      it(name, () => {
        cy.setScheduleTarget(
          {
            rules: [
              {
                weekdayFilter: [true, false, false, false, false, false, false],
                start: '15:04',
                end: '04:23',
              },
            ],
          },
          { timeZone: 'America/New_York' },
        ).then((tgt: ScheduleTarget) => {
          return cy.visit(`/schedules/${tgt.scheduleID}/assignments${params}`)
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

    check(
      'should handle selecting time values when displaying the same time zone',
      '?tz=America/New_York',
      'Sun from 3:04 PM to 4:23 AM',
    )
  })

  describe('DateTime (schedule overrides)', () => {
    const check = (name: string, params: string): Mocha.Test =>
      it(name, () => {
        cy.createSchedule({
          timeZone: 'America/New_York',
        }).then((s: Schedule) =>
          cy.visit(`/schedules/${s.id}/overrides${params}`),
        )

        if (screen === 'mobile') {
          cy.pageFab()
        } else {
          cy.get('button').contains('Create Override').click()
        }

        cy.dialogTitle('Choose Override Action')
        cy.get('[data-cy="variant.add"]').click()
        cy.dialogClick('Next')
        cy.dialogTitle('Add')

        const start = DateTime.fromJSDate(
          c.date({
            year: DateTime.utc().year + 2,
            month: DateTime.utc().month - 1,
          }) as Date,
        )

        const userName: string = (c.pickone(users) as Profile).name
        cy.dialogForm({
          addUserID: userName,
          start,
          end: start.plus({ days: 1 }),
        })

        cy.dialogFinish('Submit')

        // sanity check
        cy.get('body').contains(start.toLocaleString(DateTime.DATETIME_MED))
      })

    check(
      'should handle selecting date values when displaying the same time zone',
      '?tz=America/New_York',
    )
  })
}

testScreen('Time Pickers', testTimePickers)
