import { Chance } from 'chance'
import { testScreen } from '../support/e2e'
import { Schedule } from '../../schema'
import users from '../fixtures/users.json'

const c = new Chance()

function testSchedules(screen: ScreenFormat): void {
  describe('Creation', () => {
    it('should create a schedule', () => {
      const name = c.word({ length: 8 })
      const description = c.sentence({ words: 5 })

      cy.visit('/schedules')

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Schedule').click()
      }
      cy.dialogTitle('Create New Schedule')
      cy.dialogForm({ name, description })
      cy.dialogFinish('Submit')

      // verify on details by content headers
      cy.get('[data-cy=title]').should('contain', name)
      cy.get('[data-cy=details]').should('contain', description)
    })
  })

  describe('List Page', () => {
    it('should find a schedule', () => {
      cy.createSchedule().then((sched: Schedule) => {
        cy.visit('/schedules')
        cy.pageSearch(sched.name)
        cy.get('body')
          .should('contain', sched.name)
          .should('contain', sched.description)
      })
    })
  })

  // todo: change filters test
  describe('Details Page', () => {
    let rot: Rotation
    let sched: Schedule
    beforeEach(() => {
      cy.createRotation()
        .then((r: Rotation) => {
          rot = r
        })
        .createSchedule()
        .then((s: Schedule) => {
          sched = s
          return cy.visit('/schedules/' + sched.id)
        })
    })

    it('should delete a schedule', () => {
      cy.get('[data-cy="card-actions"]')
        .find('button[aria-label="Delete"]')
        .click()
      cy.dialogTitle('Are you sure?')
      cy.dialogFinish('Confirm')

      cy.url().should('eq', Cypress.config().baseUrl + '/schedules')

      cy.pageSearch(sched.name)
      cy.get('body').should('contain', 'No results')
      cy.reload()
      cy.get('body').should('contain', 'No results')
    })

    it('should edit a schedule', () => {
      const newName = c.word({ length: 5 })
      const newDesc = c.word({ length: 5 })
      const newTz = 'Africa/Accra'

      cy.get('[data-cy="card-actions"]')
        .find('button[aria-label="Edit"]')
        .click()
      cy.dialogTitle('Edit Schedule')
      cy.dialogForm({ name: newName, description: newDesc, 'time-zone': newTz })
      cy.dialogFinish('Submit')

      // verify changes occurred
      cy.reload()
      cy.get('[data-cy=title]').should('contain', newName)
      cy.get('[data-cy=subheader]').should('contain', newTz)
      cy.get('[data-cy=details]').should('contain', newDesc)
    })

    it('should navigate to and from assignments', () => {
      cy.navigateToAndFrom(
        screen,
        'Schedules',
        sched.name,
        'Assignments',
        `${sched.id}/assignments`,
      )
    })

    it('should navigate to and from escalation policies', () => {
      cy.navigateToAndFrom(
        screen,
        'Schedules',
        sched.name,
        'Escalation Policies',
        `${sched.id}/escalation-policies`,
      )
    })

    it('should navigate to and from overrides', () => {
      cy.navigateToAndFrom(
        screen,
        'Schedules',
        sched.name,
        'Overrides',
        `${sched.id}/overrides`,
      )
    })

    it('should navigate to and from shifts', () => {
      cy.navigateToAndFrom(
        screen,
        'Schedules',
        sched.name,
        'Shifts',
        `${sched.id}/shifts`,
      )
    })

    it('should navigate to and from on-call notifications', () => {
      cy.get('ul[data-cy="route-links"] li').should('have.lengthOf', 5)
      cy.navigateToAndFrom(
        screen,
        'Schedules',
        sched.name,
        'Notifications',
        `${sched.id}/on-call-notifications`,
      )
    })

    it('should view shifts', () => {
      cy.setScheduleTarget({
        scheduleID: sched.id,
        target: { type: 'rotation', id: rot.id },
        rules: [
          {
            start: '00:00',
            end: '00:00',
            weekdayFilter: [true, true, true, true, true, true, true],
          },
        ],
      }).then(() => {
        cy.engineTrigger()
        cy.get('[data-cy="route-links"] li').contains('Shifts').click()
        cy.get('[data-cy=flat-list-item-subheader]').should('contain', 'Today')
        cy.get('[data-cy=flat-list-item-subheader]').should(
          'contain',
          'Tomorrow',
        )
        cy.get('p').should('contain', 'Showing shifts')
      })
    })
  })

  describe('Schedule Assignments', () => {
    let rot: Rotation
    beforeEach(() => {
      cy.createRotation()
        .then((r: Rotation) => {
          rot = r
          return cy.setScheduleTarget({
            target: { id: r.id, type: 'rotation' },
          })
        })
        .then((s) => {
          cy.visit('/schedules/' + s.id + '/assignments')
          return cy.get('[role="progressbar"]').should('not.exist')
        })
    })

    it('should add a rotation as an assignment', () => {
      cy.createRotation().then(({ name }: Rotation) => {
        cy.get('ul').should('not.contain', name)

        if (screen === 'mobile') {
          cy.pageFab('Rotation')
        } else {
          cy.get('button').contains('Add Rotation').click()
        }
        cy.dialogTitle('Add Rotation to Schedule')
        cy.dialogForm({ targetID: name })
        cy.dialogFinish('Submit')

        cy.get('body').contains('li', name)
      })
    })

    it('should add a user as an assignment', () => {
      const name = rot.users[0].name

      cy.get('body').should('not.contain', name)

      if (screen === 'mobile') {
        cy.pageFab('User')
      } else {
        cy.get('button').contains('Add User').click()
      }
      cy.dialogTitle('Add User to Schedule')
      cy.dialogForm({ targetID: name })
      cy.dialogFinish('Submit')

      cy.get('body').contains('li', name)
    })

    it('should delete an assignment', () => {
      cy.get('body')
        .contains('li', rot.name)
        .find('button[data-cy=other-actions]')
        .menu('Delete')

      cy.dialogTitle('Remove Rotation')
      cy.dialogFinish('Confirm')

      cy.get('body').should('not.contain', rot.name)
    })

    it('should create multiple rules on an assignment', () => {
      if (screen === 'mobile') {
        cy.pageFab('Rotation')
      } else {
        cy.get('button').contains('Add Rotation').click()
      }
      cy.dialogTitle('Add Rotation')

      if (screen === 'mobile' || screen === 'tablet') {
        cy.dialogForm({
          targetID: rot.name,
          'rules[0].start': '02:34',
          'rules[0].end': '15:34',
        })

        cy.get('input[name="rules[0].weekdayFilter"]').siblings('div').click()
        cy.get('li').contains('Sunday').click()
        cy.focused().type('{esc}', { force: true })
      } else {
        cy.dialogForm({
          Sunday: false,
          targetID: rot.name,
          'rules[0].start': '02:34',
          'rules[0].end': '15:34',
        })
      }

      cy.get('table[data-cy="target-rules"] tbody tr').should('have.length', 1)
      cy.get('button[aria-label="Add rule"]').click()
      cy.dialogForm({ 'rules[1].start': '01:23' })
      cy.get('table[data-cy="target-rules"] tbody tr').should('have.length', 2)
      cy.dialogFinish('Submit')
      cy.get('body').should('contain', rot.name)
    })

    it('should edit an assignment', () => {
      if (screen === 'mobile' || screen === 'tablet') {
        cy.get('body')
          .contains('li', rot.name)
          .find('button[data-cy=other-actions]')
          .menu('Edit')

        cy.dialogTitle('Edit Rules')
        cy.get('input[name="rules[0].weekdayFilter"]').siblings('div').click()
        cy.get('li').contains('Wednesday').click()
        cy.focused().type('{esc}', { force: true })
        cy.dialogFinish('Submit')

        cy.get('body').contains('li', rot.name)
      } else {
        cy.get('body')
          .contains('li', rot.name)
          .find('button[data-cy=other-actions]')
          .menu('Edit')

        cy.dialogTitle('Edit Rules')
        cy.dialogForm({ Wednesday: false })
        cy.dialogFinish('Submit')

        cy.get('body').contains('li', rot.name)
      }
    })

    it('should edit then delete an assignment rule', () => {
      if (screen === 'mobile' || screen === 'tablet') {
        cy.get('body')
          .contains('li', rot.name)
          .get('button[data-cy=other-actions]')
          .menu('Edit')

        cy.dialogTitle('Edit Rules')
        cy.get('input[name="rules[0].weekdayFilter"]').siblings('div').click()
        cy.get('li').contains('Wednesday').click()
        cy.focused().type('{esc}', { force: true })

        cy.get('button[aria-label="Delete rule"]').should('not.exist')
        cy.get('button[aria-label="Add rule"').click()
        cy.get('button[aria-label="Add rule"').click()

        cy.get('button[aria-label="Delete rule"]')
          .should('have.length', 3)
          .first()
          .click()

        cy.dialogFinish('Submit')

        cy.get('body').should('contain', 'Always')
      } else {
        cy.get('body')
          .contains('li', rot.name)
          .get('button[data-cy=other-actions]')
          .menu('Edit')

        cy.dialogTitle('Edit Rules')
        cy.dialogForm({ Wednesday: true })

        cy.get('button[aria-label="Delete rule"]').should('not.exist')
        cy.get('button[aria-label="Add rule"').click()
        cy.get('button[aria-label="Add rule"').click()

        cy.get('button[aria-label="Delete rule"]')
          .should('have.length', 3)
          .first()
          .click()

        cy.dialogFinish('Submit')

        cy.get('body').should('contain', 'Always')
      }
    })
  })

  describe('Schedule Overrides', () => {
    let sched: Schedule
    beforeEach(() => {
      cy.createSchedule().then((s: Schedule) => {
        sched = s
        return cy.visit('/schedules/' + sched.id + '/overrides')
      })
    })

    it('should create an add override', () => {
      cy.get('span').should('contain', 'No results')

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Override').click()
      }

      cy.dialogTitle('Choose Override Action')
      cy.get('[data-cy="variant.add"]').click()
      cy.dialogClick('Next')
      cy.dialogTitle('Add')
      cy.dialogForm({ addUserID: users[0].name })
      cy.dialogFinish('Submit')

      cy.get('span').should('contain', users[0].name)
      cy.get('span').should('contain', 'Added from')
      expect('span').to.not.contain('No results')
    })

    it('should create a remove override', () => {
      cy.get('span').should('contain', 'No results')

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Override').click()
      }

      cy.dialogTitle('Choose Override Action')
      cy.get('[data-cy="variant.remove"]').click()
      cy.dialogClick('Next')
      cy.dialogTitle('Remove')
      cy.dialogForm({ removeUserID: users[0].name })
      cy.dialogFinish('Submit')

      cy.get('span').should('contain', users[0].name)
      cy.get('span').should('contain', 'Removed from')
      expect('span').to.not.contain('No results')
    })

    it('should create a replace override', () => {
      cy.get('span').should('contain', 'No results')

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Override').click()
      }

      cy.dialogTitle('Choose Override Action')
      cy.get('[data-cy="variant.replace"]').click()
      cy.dialogClick('Next')
      cy.dialogTitle('Replace')
      cy.dialogForm({ removeUserID: users[0].name, addUserID: users[1].name })
      cy.dialogFinish('Submit')

      cy.get('span').should('contain', users[1].name)
      cy.get('span').should('contain', `Replaces ${users[0].name} from`)
      expect('span').to.not.contain('No results')
    })

    it('should edit and delete an override', () => {
      cy.get('body').should('contain', 'No results')

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Override').click()
      }

      cy.dialogTitle('Choose Override Action')
      cy.get('[data-cy="variant.add"]').click()
      cy.dialogClick('Next')
      cy.dialogTitle('Add')
      cy.dialogForm({ addUserID: users[0].name })
      cy.dialogFinish('Submit')

      cy.get('body').should('not.contain', 'No results')
      cy.get('body').contains('li', users[0].name)

      cy.get('button[data-cy=other-actions]').menu('Edit')

      cy.dialogTitle('Edit Schedule Override')
      cy.dialogForm({ addUserID: users[1].name })
      cy.dialogFinish('Submit')

      cy.get('body')
        .should('not.contain', users[0].name)
        .should('contain', users[1].name)

      cy.get('li button[data-cy=other-actions]').menu('Delete')

      cy.dialogFinish('Confirm')

      cy.get('body').should('contain', 'No results')
    })
  })

  describe('Schedule On-Call Notifications', () => {
    let sched: Schedule

    it('should show existing notification rules', () => {
      cy.setScheduleNotificationRules(
        [
          {
            time: '00:00',
            weekdayFilter: [true, true, true, false, true, true, true],
          },
          { time: null, weekdayFilter: null },
        ],
        { timeZone: 'UTC' },
      ).then((s: Schedule) => {
        sched = s
        return cy.visit('/schedules/' + sched.id + '/on-call-notifications')
      })

      cy.get('#content')
        .should('contain', 'Notifies when on-call changes')
        .should('contain', 'Notifies Sun—Tue, Thu—Sat at 12:00 AM')
    })

    it('should create notification rules', () => {
      cy.createSchedule({ timeZone: 'UTC' }).then((s: Schedule) => {
        sched = s
        return cy.visit('/schedules/' + sched.id + '/on-call-notifications')
      })

      // on change
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Notification Rule').click()
      }
      cy.dialogTitle('Create Notification Rule')
      cy.dialogForm({
        ruleType: 'on-change',
        'slack-channel-id': 'general',
      })
      cy.dialogFinish('Submit')
      cy.get('body').should('contain', '#general')
      cy.get('body').should('contain', 'Notifies when on-call changes')

      // time of day
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Notification Rule').click()
      }
      cy.dialogTitle('Create Notification Rule')
      cy.dialogForm({
        ruleType: 'time-of-day',
        time: '00:00',
        'weekdayFilter[0]': false,
        'weekdayFilter[1]': true,
        'weekdayFilter[2]': false,
        'weekdayFilter[3]': false,
        'weekdayFilter[4]': false,
        'weekdayFilter[5]': false,
        'weekdayFilter[6]': false,
        'slack-channel-id': 'foobar',
      })
      cy.dialogFinish('Submit')
      cy.get('#content').should('contain', 'Notifies Mon at 12:00 AM')
    })

    it('should delete a notification rule', () => {
      cy.setScheduleNotificationRules(
        [
          {
            time: '00:00',
            weekdayFilter: [true, true, true, false, true, true, true],
          },
          { time: null, weekdayFilter: null },
        ],
        { timeZone: 'UTC' },
      ).then((s: Schedule) => {
        sched = s
        return cy.visit('/schedules/' + sched.id + '/on-call-notifications')
      })

      cy.get('#content')
        .contains('li', 'on-call changes')
        .find('[aria-label="Other Actions"]')
        .menu('Delete')

      cy.dialogTitle('Are you sure?')
      cy.dialogContains(' will no longer be notified when on-call changes.')
      cy.dialogFinish('Confirm')

      cy.get('#content').should('not.contain', 'on-call changes')

      cy.get('#content')
        .contains('li', 'Notifies')
        .find('[aria-label="Other Actions"]')
        .menu('Delete')

      cy.dialogTitle('Are you sure?')
      cy.dialogContains(
        ' will no longer be notified Sun—Tue, Thu—Sat at 12:00 AM',
      )
      cy.dialogFinish('Confirm')

      cy.get('#content')
        .should('not.contain', 'Notifies')
        .should('contain', 'No notification rules')
    })

    it('should edit from onSchedule to onChange', () => {
      cy.setScheduleNotificationRules(
        [
          {
            time: '00:00',
            weekdayFilter: [false, true, true, true, true, true, true],
          },
        ],
        { timeZone: 'UTC' },
      ).then((s: Schedule) => {
        sched = s
        return cy.visit('/schedules/' + sched.id + '/on-call-notifications')
      })

      cy.get('#content')
        .contains('li', 'Notifies')
        .find('[aria-label="Other Actions"]')
        .menu('Edit')

      cy.dialogTitle('Edit Notification Rule')
      cy.dialogForm({ ruleType: 'on-change' })
      cy.dialogFinish('Submit')
      cy.get('#content').should('contain', 'Notifies when on-call changes')
    })

    it('should edit from onChange to onSchedule', () => {
      cy.setScheduleNotificationRules([{ time: null, weekdayFilter: null }], {
        timeZone: 'UTC',
      }).then((s: Schedule) => {
        sched = s
        return cy.visit('/schedules/' + sched.id + '/on-call-notifications')
      })

      cy.get('#content')
        .contains('li', 'Notifies')
        .find('[aria-label="Other Actions"]')
        .menu('Edit')

      cy.dialogTitle('Edit Notification Rule')
      cy.dialogForm({
        ruleType: 'time-of-day',
        time: '07:00',
        'weekdayFilter[0]': false,
        'weekdayFilter[1]': true,
        'weekdayFilter[2]': false,
        'weekdayFilter[3]': false,
        'weekdayFilter[4]': false,
        'weekdayFilter[5]': false,
        'weekdayFilter[6]': false,
        'slack-channel-id': 'foobar',
      })
      cy.dialogFinish('Submit')
      cy.get('body').should('contain', 'Notifies Mon at 7:00 AM')
    })
  })
}

testScreen('Schedules', testSchedules)

testScreen('Slack User Group Support', (screen: ScreenFormat) => {
  describe('Schedule On-Call Notifications', () => {
    let sched: Schedule
    it('should create notification rules with slack user groups', () => {
      cy.createSchedule({ timeZone: 'UTC' }).then((s: Schedule) => {
        sched = s
        return cy.visit('/schedules/' + sched.id + '/on-call-notifications')
      })

      // on change
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Notification Rule').click()
      }

      cy.dialogTitle('Create Notification Rule')
      cy.dialogForm({
        ruleType: 'on-change',
        'dest.type': 'Update Slack User Group',
        'slack-usergroup-id': 'foobar',
        'slack-channel-id': 'foobar',
      })

      cy.dialogFinish('Submit')
      cy.get('body').should('contain', '@foobar')
      cy.get('body').should('contain', 'Notifies when on-call changes')

      // time of day
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Notification Rule').click()
      }
      cy.dialogTitle('Create Notification Rule')
      cy.dialogForm({
        ruleType: 'time-of-day',
        time: '00:00',
        'weekdayFilter[0]': false,
        'weekdayFilter[1]': true,
        'weekdayFilter[2]': false,
        'weekdayFilter[3]': false,
        'weekdayFilter[4]': false,
        'weekdayFilter[5]': false,
        'weekdayFilter[6]': false,
        'dest.type': 'Update Slack User Group',
        'slack-usergroup-id': 'foobar',
        'slack-channel-id': 'foobar',
      })
      cy.dialogFinish('Submit')
      cy.get('#content').should('contain', 'Notifies Mon at 12:00 AM')
    })

    it('should create notification rules with webhook', () => {
      cy.updateConfig({ Webhook: { Enable: true } })
      cy.reload()
      cy.createSchedule({ timeZone: 'UTC' }).then((s: Schedule) => {
        sched = s
        return cy.visit('/schedules/' + sched.id + '/on-call-notifications')
      })

      // on change
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Notification Rule').click()
      }

      cy.dialogTitle('Create Notification Rule')
      cy.dialogForm({
        ruleType: 'on-change',
        'dest.type': 'Webhook',
        webhook_url: 'http://www.example.com',
      })

      cy.dialogFinish('Submit')
      cy.get('body').should('contain', 'www.example.com')
      cy.get('body').should('contain', 'Notifies when on-call changes')

      // time of day
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Notification Rule').click()
      }
      cy.dialogTitle('Create Notification Rule')
      cy.dialogForm({
        ruleType: 'time-of-day',
        'dest.type': 'Webhook',
        webhook_url: 'http://www.example.com',
        time: '00:00',
        'weekdayFilter[0]': false,
        'weekdayFilter[1]': true,
        'weekdayFilter[2]': false,
        'weekdayFilter[3]': false,
        'weekdayFilter[4]': false,
        'weekdayFilter[5]': false,
        'weekdayFilter[6]': false,
      })
      cy.dialogFinish('Submit')
      cy.get('#content').should('contain', 'Notifies Mon at 12:00 AM')
    })
  })
})
