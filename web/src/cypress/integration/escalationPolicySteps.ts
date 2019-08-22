import { Chance } from 'chance'
import { testScreen } from '../support'

const c = new Chance()

testScreen('Escalation Policy Steps', testSteps)

function testSteps(screen: ScreenFormat) {
  describe('Steps', () => {
    let ep: EP
    let r1: Rotation
    let r2: Rotation
    let s1: Schedule
    let s2: Schedule

    beforeEach(() => {
      cy.createRotation().then(r => (r1 = r))
      cy.createRotation().then(r => (r2 = r))
      cy.createSchedule().then(s => (s1 = s))
      cy.createSchedule().then(s => (s2 = s))

      cy.createEP().then(e => {
        ep = e
        return cy.visit(`/escalation-policies/${ep.id}`)
      })
    })

    it('should see no steps text', () => {
      cy.get('body').should(
        'contain',
        'No steps currently on this Escalation Policy',
      )
    })

    // Create a step with 2 of each type of GoAlert target
    it('should create a step', () => {
      cy.fixture('users').then(users => {
        const u1 = users[0]
        const u2 = users[1]

        cy.pageFab()
        cy.get('div[role=dialog]').as('dialog')
        cy.get('@dialog').should('contain', 'Create Step')

        cy.get('input[name=rotations]').selectByLabel(r1.name)
        cy.get('input[name=rotations]').selectByLabel(r2.name)

        cy.get('button[data-cy="schedules-step"]').click()
        cy.get('input[name=schedules]').selectByLabel(s1.name)
        cy.get('input[name=schedules]').selectByLabel(s2.name)

        cy.get('button[data-cy="users-step"]').click()
        cy.get('input[name=users]').selectByLabel(u1.name)
        cy.get('input[name=users]').selectByLabel(u2.name)
        const del = c.integer({ min: 1, max: 9000 })
        const delStr = del.toString()
        cy.get('input[name=delayMinutes]')
          .clear()
          .type(delStr)

        // submit form
        cy.get('button[type=submit]').click()

        // confirm dialog closes
        cy.get('@dialog').should('not.exist')

        // verify data integrity
        cy.get('body').should('contain', 'Notify the following:')
        cy.get('body').should('contain', 'Step #1:')
        cy.get('div[data-cy=rotation-chip]').should('contain', r1.name)
        cy.get('div[data-cy=rotation-chip]').should('contain', r2.name)
        cy.get('div[data-cy=schedule-chip]').should('contain', s1.name)
        cy.get('div[data-cy=schedule-chip]').should('contain', s2.name)
        cy.get('div[data-cy=user-chip]').should('contain', u1.name)
        cy.get('div[data-cy=user-chip]').should('contain', u2.name)
        cy.get('body').should(
          'contain',
          `Go back to step #1 after ${delStr} minutes`,
        )
      })
    })

    it('should add users when slack is disabled', () => {
      cy.updateConfig({ Slack: { Enable: false } })
      cy.reload()
      cy.fixture('users').then(users => {
        const u1 = users[0]
        const u2 = users[1]

        cy.pageFab()
        cy.get('div[role=dialog]').as('dialog')
        cy.get('@dialog').should('contain', 'Create Step')

        cy.get('button[data-cy="users-step"]').click()
        cy.get('input[name=users]').selectByLabel(u1.name)
        cy.get('input[name=users]').selectByLabel(u2.name)
      })
    })

    it('should edit a step', () => {
      cy.createEPStep({ epID: ep.id }).then(() => cy.reload())
      cy.get('ul[data-cy=steps-list]')
        .find('li')
        .eq(0)
        .find('button[data-cy=other-actions]')
        .menu('Edit')

      cy.get('div[role=dialog]').as('dialog')
      cy.get('@dialog').should('contain', 'Edit Step')

      cy.get('input[name=rotations]').selectByLabel(r1.name)

      const del = c.integer({ min: 1, max: 9000 })
      const delStr = del.toString()
      cy.get('input[name=delayMinutes]')
        .clear()
        .type(delStr)

      // submit form
      cy.get('button[type=submit]').click()

      // confirm dialog closes
      cy.get('@dialog').should('not.exist')

      // verify data integrity
      cy.get('body').should('contain', 'Notify the following:')
      cy.get('body').should('contain', 'Step #1:')
      cy.get('div[data-cy=rotation-chip]').should('contain', r1.name)
      cy.get('body').should(
        'contain',
        `Go back to step #1 after ${delStr} minutes`,
      )
    })

    it('should add and then remove a slack channel', () => {
      cy.updateConfig({ Slack: { Enable: true } })
      cy.reload()

      cy.pageFab()
      cy.get('div[role=dialog]').as('dialog')
      cy.get('@dialog').should('contain', 'Create Step')

      // expand slack channels section
      cy.get('button[data-cy="slack-channels-step"]').click()

      // add slack channels
      cy.get('input[name=slackChannels]').selectByLabel('general')
      cy.get('input[name=slackChannels]').selectByLabel('foobar')

      // submit create form
      cy.get('button[type=submit]').click()

      // confirm create dialog closes
      cy.get('@dialog').should('not.exist')

      // verify data integrity
      cy.get('body').should('contain', 'Notify the following:')
      cy.get('body').should('contain', 'Step #1:')
      cy.get('div[data-cy=slack-chip]').should('contain', '#general')
      cy.get('div[data-cy=slack-chip]').should('contain', '#foobar')

      // open edit step dialog
      cy.get('ul[data-cy=steps-list]')
        .find('li')
        .eq(0)
        .find('button[data-cy=other-actions]')
        .menu('Edit')

      // confirm edit step dialog open
      cy.get('@dialog').should('contain', 'Edit Step')

      // expand slack channels section
      cy.get('button[data-cy="slack-channels-step"]').click()

      // delete foobar channel
      cy.get('input[name=slackChannels]').multiRemoveByLabel('#foobar')

      // submit edit form
      cy.get('button[type=submit]').click()

      // confirm edit dialog closes
      cy.get('@dialog').should('not.exist')

      // verify data integrity
      cy.get('body').should('contain', 'Notify the following:')
      cy.get('body').should('contain', 'Step #1:')
      cy.get('div[data-cy=slack-chip]').should('contain', '#general')
      cy.get('div[data-cy=slack-chip]').should('not.contain', '#foobar')
    })

    it('should delete a step', () => {
      cy.createEPStep({ epID: ep.id }).then(() => cy.reload())
      cy.get('ul[data-cy=steps-list]')
        .find('li')
        .eq(0)
        .find('button[data-cy=other-actions]')
        .menu('Delete')
      cy.get('div[role=dialog]').as('dialog')

      cy.get('@dialog').should('contain', 'Are you sure?')
      cy.get('@dialog').should(
        'contain',
        'This will delete step #1 on this escalation policy.',
      )
      cy.get('button[type=submit]').click()

      cy.get('@dialog').should('not.exist')
      cy.get('body').should(
        'contain',
        'No steps currently on this Escalation Policy',
      )
    })

    it('should drag and drop a step', () => {
      let s1: EPStep
      let s2: EPStep
      let s3: EPStep

      cy.createEPStep({ epID: ep.id })
        .then(x => {
          s1 = x
          cy.createEPStep({ epID: ep.id }).then(y => {
            s2 = y
            cy.createEPStep({ epID: ep.id }).then(z => {
              s3 = z
              cy.reload()
            })
          })
        })
        .then(() => {
          cy.get('ul[data-cy=steps-list]')
            .find('li')
            .should('have.length', 3)

          // focus element to be drag and dropped
          cy.get('ul[data-cy=steps-list]')
            .find('li')
            .eq(0)
            .should('contain', 'Step #1')
            .should('contain', s1.delayMinutes)
            .parent('[tabindex]')
            .focus()
            .type(' ')

          cy.get('body').should(
            'contain',
            'You have lifted an item in position 1',
          )

          // move element down one position
          cy.focused().type('{downarrow}', { force: true })

          cy.get('body')
            .should('contain', 'You have moved the item from position 1')
            .should('contain', 'to position 2')

          // move element down one more position
          cy.focused().type('{downarrow}', { force: true })

          cy.get('body')
            .should('contain', 'You have moved the item from position 1')
            .should('contain', 'to position 3')

          // drop element
          cy.focused().type(' ', { force: true })

          // verify data integrity
          cy.get('ul[data-cy=steps-list] :nth-child(1) li')
            .should('contain', 'Step #1')
            .should('contain', s2.delayMinutes)
          cy.get('ul[data-cy=steps-list] :nth-child(2) li')
            .should('contain', 'Step #2')
            .should('contain', s3.delayMinutes)
          cy.get('ul[data-cy=steps-list] :nth-child(3) li')
            .should('contain', 'Step #3')
            .should('contain', s1.delayMinutes)
        })
    })
  })
}
