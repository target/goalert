import { Chance } from 'chance'
import { testScreen } from '../support'
const c = new Chance()

testScreen('Schedules', testSchedules)

function testSchedules(screen: ScreenFormat) {
  describe('Creation', () => {
    it('should create a schedule', () => {
      const name = c.word({ length: 8 })
      const description = c.sentence({ words: 5 })

      cy.visit('/schedules')

      cy.pageFab()
      cy.dialogTitle('Create New Schedule')
      cy.dialogForm({ name, description })
      cy.dialogFinish('Submit')

      // verify on details by content headers
      cy.get('[data-cy=details-heading]').should('contain', name)
      cy.get('[data-cy=details]').should('contain', description)
    })
  })

  describe('List Page', () => {
    it('should find a schedule', () => {
      cy.createSchedule().then(sched => {
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
        .then(r => {
          rot = r
        })
        .createSchedule()
        .then(s => {
          sched = s
          return cy.visit('/schedules/' + sched.id)
        })
    })

    it('should delete a schedule', () => {
      cy.pageAction('Delete Schedule')
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

      cy.pageAction('Edit Schedule')
      cy.dialogTitle('Edit Schedule')
      cy.dialogForm({ name: newName, description: newDesc, 'time-zone': newTz })
      cy.dialogFinish('Submit')

      // verify changes occurred
      cy.reload()
      cy.get('[data-cy=details-heading]').should('contain', newName)
      cy.get('[data-cy=details]').should('contain', newDesc)
      cy.get('[data-cy=title-footer]').should('contain', newTz)
    })

    it('should navigate to and from assignments', () => {
      cy.navigateToAndFrom(
        screen,
        'Schedule Details',
        sched.name,
        'Assignments',
        `${sched.id}/assignments`,
      )
    })

    it('should navigate to and from escalation policies', () => {
      cy.navigateToAndFrom(
        screen,
        'Schedule Details',
        sched.name,
        'Escalation Policies',
        `${sched.id}/escalation-policies`,
      )
    })

    it('should navigate to and from overrides', () => {
      cy.navigateToAndFrom(
        screen,
        'Schedule Details',
        sched.name,
        'Overrides',
        `${sched.id}/overrides`,
      )
    })

    it('should navigate to and from shifts', () => {
      cy.navigateToAndFrom(
        screen,
        'Schedule Details',
        sched.name,
        'Shifts',
        `${sched.id}/shifts`,
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
        cy.get('li')
          .contains('Shifts')
          .click()
        cy.reload()
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
    let sched: ScheduleTarget
    beforeEach(() => {
      cy.createRotation()
        .then(r => {
          rot = r
          return cy.setScheduleTarget({
            target: { id: r.id, type: 'rotation' },
          })
        })
        .then(s => {
          sched = s
          return cy.visit('/schedules/' + sched.schedule.id + '/assignments')
        })
    })

    it('should add a rotation as an assignment', () => {
      const name = rot.name

      cy.get('body').should('not.contain', name)

      cy.pageFab('Rotation')
      cy.dialogTitle('Add Rotation to Schedule')
      cy.dialogForm({ targetID: name })
      cy.dialogFinish('Submit')

      cy.get('body').contains('li', name)
    })

    it('should add a user as an assignment', () => {
      const name = rot.users[0].name

      cy.get('body').should('not.contain', name)

      cy.pageFab('User')
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
      // todo: mobile dialog is completely different
      if (screen === 'mobile' || screen === 'tablet') return

      cy.pageFab('Rotation')

      cy.dialogTitle('Add Rotation')
      cy.dialogForm({
        Sunday: false,
        targetID: rot.name,
        'rules[0].start': '02:34',
        'rules[0].end': '15:34',
      })

      cy.get('table[data-cy="target-rules"] tbody tr').should('have.length', 1)

      cy.get('button[aria-label="Add rule"]').click()

      cy.dialogForm({ 'rules[1].start': '01:23' })

      cy.get('table[data-cy="target-rules"] tbody tr').should('have.length', 2)

      cy.dialogFinish('Submit')

      cy.get('body').should('contain', rot.name)
    })

    it('should edit an assignment', () => {
      // todo: mobile dialog is completely different
      if (screen === 'mobile' || screen === 'tablet') return

      cy.get('body')
        .contains('li', rot.name)
        .find('button[data-cy=other-actions]')
        .menu('Edit')

      cy.dialogTitle('Edit Rules')
      cy.dialogForm({ Wednesday: false })
      cy.dialogFinish('Submit')

      cy.get('body').contains('li', rot.name)
    })

    it('should edit then delete an assignment rule', () => {
      // todo: mobile dialog is completely different
      if (screen === 'mobile' || screen === 'tablet') return

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
    })
  })

  describe('Schedule Overrides', () => {
    let sched: Schedule
    beforeEach(() => {
      cy.createSchedule().then(s => {
        sched = s
        return cy.visit('/schedules/' + sched.id + '/overrides')
      })
    })

    it('should create an add override', () => {
      cy.fixture('users').then(users => {
        cy.get('span').should('contain', 'No results')

        cy.pageFab('Add')
        cy.dialogTitle('Add a User')
        cy.dialogForm({ addUserID: users[0].name })
        cy.dialogFinish('Submit')

        cy.get('span').should('contain', users[0].name)
        cy.get('p').should('contain', 'Added from')
        expect('span').to.not.contain('No results')
      })
    })

    it('should create a remove override', () => {
      cy.fixture('users').then(users => {
        cy.get('span').should('contain', 'No results')

        cy.pageFab('Remove')
        cy.dialogTitle('Remove a User')
        cy.dialogForm({ removeUserID: users[0].name })
        cy.dialogFinish('Submit')

        cy.get('span').should('contain', users[0].name)
        cy.get('p').should('contain', 'Removed from')
        expect('span').to.not.contain('No results')
      })
    })

    it('should create a replace override', () => {
      cy.fixture('users').then(users => {
        cy.get('span').should('contain', 'No results')

        cy.pageFab('Replace')
        cy.dialogTitle('Replace a User')
        cy.dialogForm({ removeUserID: users[0].name, addUserID: users[1].name })
        cy.dialogFinish('Submit')

        cy.get('span').should('contain', users[1].name)
        cy.get('p').should('contain', `Replaces ${users[0].name} from`)
        expect('span').to.not.contain('No results')
      })
    })

    it('should edit and delete an override', () => {
      cy.fixture('users').then(users => {
        cy.get('body').should('contain', 'No results')

        cy.pageFab('Add')
        cy.dialogTitle('Add a User')
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
  })
}
