import { Chance } from 'chance'
const c = new Chance()
import { testScreen } from '../support'

testScreen('Schedules', testSchedules)

function testSchedules(screen: ScreenFormat) {
  describe('Creation', () => {
    it('should create a schedule', () => {
      const name = c.word({ length: 8 })
      const description = c.sentence({ words: 5 })

      cy.visit('/schedules')
      cy.pageFab()
      cy.get('input[name=name]').type(name)

      cy.get('textarea[name=description]')
        .clear()
        .type(description)
      cy.get('button')
        .contains('Submit')
        .click()

      // verify on details by content headers
      cy.get('[data-cy=details-heading]').contains(name)
      cy.get('[data-cy=details]').contains(description)
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
      cy.get('button')
        .contains('Confirm')
        .click()

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
      cy.get('input[name=name]')
        .clear()
        .type(newName) // type in new name
      cy.get('input[name=time-zone]').selectByLabel(newTz)

      cy.get('textarea[name=description]')
        .clear()
        .type(newDesc) // type in new description
      cy.get('button')
        .contains('Submit')
        .click()

      // verify changes occurred
      cy.reload()
      cy.get('[data-cy=details-heading]').contains(newName)
      cy.get('[data-cy=details]').contains(newDesc)
      cy.get('[data-cy=title-footer]').contains(newTz)
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
      }).then(tgt => {
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
      cy.pageFab('Rotation')

      // select create rotation
      cy.get('input[name=targetID]').selectByLabel(rot.name)
      cy.get('button')
        .contains('Submit')
        .click()

      cy.get('body').contains('li', rot.name)
    })

    it('should delete an assignment', () => {
      cy.get('body')
        .contains('li', rot.name)
        .find('button[data-cy=other-actions]')
        .menu('Delete')

      cy.get('body')
        .contains('button', 'Confirm')
        .click()

      cy.get('body').should('not.contain', rot.name)
    })

    it('should edit an assignment', () => {
      // todo: mobile dialog is completely different
      if (screen === 'mobile' || screen === 'tablet') return

      cy.get('body')
        .contains('li', rot.name)
        .find('button[data-cy=other-actions]')
        .menu('Edit')

      cy.get('input[name=Wednesday]').click()
      cy.get('body')
        .contains('button', 'Submit')
        .click()
      cy.get('body').contains('li', rot.name)
    })

    it('should edit then delete an assignment rule', () => {
      // todo: mobile dialog is completely different
      if (screen === 'mobile' || screen === 'tablet') return

      cy.get('body')
        .contains('li', rot.name)
        .get('button[data-cy=other-actions]')
        .menu('Edit')

      cy.get('input[name=Wednesday]').click()
      cy.get('button[aria-label="Delete rule"]').should('not.exist')
      cy.get('button[aria-label="Add rule"').click()
      cy.get('button[aria-label="Add rule"').click()

      cy.get('button[aria-label="Delete rule"]')
        .should('have.length', 3)
        .first()
        .click()
      cy.get('body')
        .contains('button', 'Submit')
        .click()
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

        cy.get('input[name=addUserID]').selectByLabel(users[0].name)
        cy.get('button')
          .contains('Submit')
          .click()

        cy.get('span').should('contain', users[0].name)
        cy.get('p').should('contain', 'Added from')
        expect('span').to.not.contain('No results')
      })
    })

    it('should create a remove override', () => {
      cy.fixture('users').then(users => {
        cy.get('span').should('contain', 'No results')

        cy.pageFab('Remove')

        cy.get('input[name=removeUserID]').selectByLabel(users[0].name)
        cy.get('button')
          .contains('Submit')
          .click()

        cy.get('span').should('contain', users[0].name)
        cy.get('p').should('contain', 'Removed from')
        expect('span').to.not.contain('No results')
      })
    })

    it('should create a replace override', () => {
      cy.fixture('users').then(users => {
        cy.get('span').should('contain', 'No results')

        cy.pageFab('Replace')

        cy.get('input[name=removeUserID]').selectByLabel(users[0].name)
        cy.get('input[name=addUserID]').selectByLabel(users[1].name)

        cy.get('button')
          .contains('Submit')
          .click()

        cy.get('span').should('contain', users[1].name)
        cy.get('p').should('contain', `Replaces ${users[0].name} from`)
        expect('span').to.not.contain('No results')
      })
    })

    it('should edit an override', () => {
      cy.fixture('users').then(users => {
        cy.get('body').should('contain', 'No results')

        cy.pageFab('Add')

        cy.get('input[name=addUserID]').selectByLabel(users[0].name)
        cy.get('button')
          .contains('Submit')
          .click()

        cy.get('body').should('not.contain', 'No results')

        cy.get('body').should('contain', users[0].name)

        cy.get('button[data-cy=other-actions]').menu('Edit')

        cy.get('input[name=addUserID]').selectByLabel(users[1].name)
        cy.get('button')
          .contains('Submit')
          .click()

        cy.get('body')
          .should('not.contain', users[0].name)
          .should('contain', users[1].name)

        cy.get('button[data-cy=other-actions]').menu('Delete')

        cy.get('button')
          .contains('Confirm')
          .click()

        cy.get('body').should('contain', 'No results')
      })
    })
  })
}
