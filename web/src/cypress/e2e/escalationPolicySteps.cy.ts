import { Chance } from 'chance'
import { testScreen } from '../support/e2e'
import { Schedule } from '../../schema'
import users from '../fixtures/users.json'

const c = new Chance()

function testSteps(screen: ScreenFormat): void {
  describe('Steps', () => {
    let ep: EP
    let r1: Rotation
    let r2: Rotation
    let s1: Schedule
    let s2: Schedule

    beforeEach(() => {
      cy.createRotation().then((r: Rotation) => (r1 = r))
      cy.createRotation().then((r: Rotation) => (r2 = r))
      cy.createSchedule().then((s: Schedule) => (s1 = s))
      cy.createSchedule().then((s: Schedule) => (s2 = s))

      cy.createEP().then((e: EP) => {
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
      const u1 = users[0]
      const u2 = users[1]
      const delay = c.integer({ min: 1, max: 9000 })

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Step').click()
      }
      cy.dialogTitle('Create Step')
      cy.dialogForm({ 'dest.type': 'Schedule', schedule_id: s1.name })
      cy.dialogClick('Add Destination')
      cy.get('div[role="dialog"] [data-testid=destination-chip]').should(
        'contain',
        s1.name,
      )
      cy.dialogForm({ schedule_id: s2.name })
      cy.dialogClick('Add Destination')
      cy.get('div[role="dialog"] [data-testid=destination-chip]').should(
        'contain',
        s2.name,
      )

      cy.dialogForm({ 'dest.type': 'User', user_id: u1.name })
      cy.dialogClick('Add Destination')
      cy.get('div[role="dialog"] [data-testid=destination-chip]').should(
        'contain',
        u1.name,
      )
      cy.dialogForm({ user_id: u2.name })
      cy.dialogClick('Add Destination')
      cy.get('div[role="dialog"] [data-testid=destination-chip]').should(
        'contain',
        u2.name,
      )

      cy.dialogForm({ 'dest.type': 'Rotation', rotation_id: r1.name })
      cy.dialogClick('Add Destination')
      cy.get('div[role="dialog"] [data-testid=destination-chip]').should(
        'contain',
        r1.name,
      )
      cy.dialogForm({ rotation_id: r2.name })
      cy.dialogClick('Add Destination')
      cy.get('div[role="dialog"] [data-testid=destination-chip]').should(
        'contain',
        r2.name,
      )

      cy.dialogForm({ delayMinutes: delay.toString() })
      cy.dialogFinish('Submit')

      // verify data integrity
      cy.get('body').should('contain', 'Notify the following:')
      cy.get('body').should('contain', 'Step #1:')
      cy.get('[data-testid=destination-chip]').should('contain', r1.name)
      cy.get('[data-testid=destination-chip]').should('contain', r2.name)
      cy.get('[data-testid=destination-chip]').should('contain', s1.name)
      cy.get('[data-testid=destination-chip]').should('contain', s2.name)
      cy.get('[data-testid=destination-chip]').should('contain', u1.name)
      cy.get('[data-testid=destination-chip]').should('contain', u2.name)
      cy.get('body').should(
        'contain',
        `Go back to step #1 after ${delay.toString()} minutes`,
      )
    })

    it('should add users when slack is disabled', () => {
      cy.updateConfig({ Slack: { Enable: false } })
      cy.reload()
      const u1 = users[0]
      const u2 = users[1]

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Step').click()
      }
      cy.dialogTitle('Create Step')

      cy.dialogForm({ 'dest.type': 'User', user_id: u1.name })
      cy.dialogClick('Add Destination')
      cy.dialogForm({ user_id: u2.name })
      cy.dialogClick('Add Destination')
    })

    it('should edit a step', () => {
      cy.createEPStep({ epID: ep.id })
        .then(() => {
          cy.reload()
        })
        .then(() => {
          const delay = c.integer({ min: 1, max: 9000 })

          cy.get('ul[data-cy=steps-list] :nth-child(2) li')
            .should('contain', 'Step #')
            .find('button[data-cy=other-actions]')
            .menu('Edit')

          cy.dialogTitle('Edit Step')
          cy.dialogForm({
            'dest.type': 'Schedule',
            schedule_id: s1.name,
            delayMinutes: delay.toString(),
          })
          cy.dialogClick('Add Destination')
          cy.get('div[role="dialog"] [data-testid=destination-chip]').should(
            'contain',
            s1.name,
          )

          cy.dialogFinish('Submit')

          // verify data integrity
          cy.get('body').should('contain', 'Notify the following:')
          cy.get('body').should('contain', 'Step #1:')
          cy.get('[data-testid=destination-chip]').should('contain', s1.name)
          cy.get('body').should(
            'contain',
            `Go back to step #1 after ${delay.toString()} minutes`,
          )
        })
    })

    it('should add, click, and remove a slack channel', () => {
      cy.updateConfig({ Slack: { Enable: true } })
      cy.reload()

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Step').click()
      }
      cy.dialogTitle('Create Step')

      // add slack channels
      cy.dialogForm({
        'dest.type': 'Slack Channel',
        slack_channel_id: 'general',
      })
      cy.dialogClick('Add Destination')
      cy.get('div[role="dialog"] [data-testid=destination-chip]').should(
        'contain',
        '#general',
      )

      cy.dialogForm({ slack_channel_id: 'foobar' })
      cy.dialogClick('Add Destination')
      cy.get('div[role="dialog"] [data-testid=destination-chip]').should(
        'contain',
        '#foobar',
      )

      cy.dialogFinish('Submit')

      // verify data integrity
      cy.get('body').should('contain', 'Notify the following:')
      cy.get('body').should('contain', 'Step #1:')
      cy.get('[data-testid=destination-chip]').should('contain', '#general')
      cy.get('[data-testid=destination-chip]').should('contain', '#foobar')

      // verify clickability
      cy.window().then((win) => {
        cy.stub(win, 'open').as('slackRedirect')
      })
      cy.get('a[data-testid=destination-chip]').should(
        'have.attr',
        'target',
        '_blank',
      )

      // open edit step dialog
      cy.get('ul[data-cy=steps-list] :nth-child(2) li')
        .find('button[data-cy=other-actions]')
        .menu('Edit')

      cy.dialogTitle('Edit Step')

      // delete foobar channel
      cy.get('[role=dialog] a[data-testid=destination-chip]')
        .contains('a', '#foobar')
        .find('[data-testid=CancelIcon]')
        .click()
      cy.get('div[role="dialog"] [data-testid=destination-chip]')
        .contains('#foobar')
        .should('not.exist')

      cy.dialogFinish('Submit')

      // verify data integrity
      cy.get('body').should('contain', 'Notify the following:')
      cy.get('body').should('contain', 'Step #1:')
      cy.get('[data-testid=destination-chip]').should('contain', '#general')
      cy.get('[data-testid=destination-chip]').should('not.contain', '#foobar')
    })

    it('should delete a step', () => {
      cy.createEPStep({ epID: ep.id }).then(() => cy.reload())
      cy.get('ul[data-cy=steps-list] :nth-child(2) li')
        .find('button[data-cy=other-actions]')
        .menu('Delete')
      cy.dialogTitle('Are you sure?')
      cy.dialogContains('This will delete step #1 on this escalation policy.')
      cy.dialogFinish('Confirm')

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
        .then((x: EPStep) => {
          s1 = x
          cy.createEPStep({ epID: ep.id }).then((y: EPStep) => {
            s2 = y
            cy.createEPStep({ epID: ep.id }).then((z: EPStep) => {
              s3 = z
              cy.reload()
            })
          })
        })
        .then(() => {
          // length of 4 = 3 steps + 1 subheader
          cy.get('ul[data-cy=steps-list]')
            .should('contain', 'Step #3')
            .find('li')
            .should('have.length', 4)

          // pick up first step
          cy.get('[id="drag-0"]').focus()
          cy.focused().type('{enter}')
          cy.get('body').should(
            'contain',
            'Picked up sortable item 1. Sortable item 1 is in position 1 of 3',
          )

          // re-order
          cy.focused().type('{downarrow}', { force: true })
          cy.get('body').should(
            'contain',
            'Sortable item 1 was moved into position 2 of 3',
          )

          // place step, calls mutation
          cy.focused().type('{enter}', { force: true })
          cy.get('body').should(
            'contain',
            'Sortable item 1 was dropped at position 2 of 3',
          )

          // verify re-order
          cy.get('ul[data-cy=steps-list] :nth-child(2) li')
            .should('contain', 'Step #1')
            .should('contain', s2.delayMinutes)
          cy.get('ul[data-cy=steps-list] :nth-child(3) li')
            .should('contain', 'Step #2')
            .should('contain', s1.delayMinutes)
          cy.get('ul[data-cy=steps-list] :nth-child(4) li')
            .should('contain', 'Step #3')
            .should('contain', s3.delayMinutes)
        })
    })
  })
}

testScreen('Escalation Policy Steps', testSteps)

testScreen('Webhook Support', (screen: ScreenFormat) => {
  let ep: EP
  beforeEach(() => {
    cy.createEP().then((e: EP) => {
      ep = e
      cy.visit(`/escalation-policies/${ep.id}`)
    })
  })

  it('should add, click, and remove a webhook', () => {
    cy.updateConfig({ Webhook: { Enable: true } })
    cy.reload()

    if (screen === 'mobile') {
      cy.pageFab()
    } else {
      cy.get('button').contains('Create Step').click()
    }
    cy.dialogTitle('Create Step')

    // add webhooks
    cy.dialogForm({
      'dest.type': 'Webhook',
      webhook_url: 'https://webhook.site',
    })
    cy.dialogClick('Add Destination')
    cy.get('div[role="dialog"] [data-testid=destination-chip]').should(
      'contain',
      'webhook.site',
    )

    cy.dialogForm({
      webhook_url: 'https://example.com',
    })
    cy.dialogClick('Add Destination')
    cy.get('div[role="dialog"] [data-testid=destination-chip]').should(
      'contain',
      'example.com',
    )

    cy.dialogFinish('Submit')

    // verify data integrity
    cy.get('body').should('contain', 'Notify the following:')
    cy.get('body').should('contain', 'Step #1:')
    cy.get('[data-testid=destination-chip]').should('contain', 'webhook.site')
    cy.get('[data-testid=destination-chip]').should('contain', 'example.com')

    // open edit step dialog
    cy.get('ul[data-cy=steps-list] :nth-child(2) li')
      .find('button[data-cy=other-actions]')
      .menu('Edit')

    cy.dialogTitle('Edit Step')

    // delete webhook.site webhook
    cy.get('[role=dialog] [data-testid=destination-chip]')
      .contains('[data-testid=destination-chip]', 'webhook.site')
      .find('[data-testid=CancelIcon]')
      .click()

    cy.dialogFinish('Submit')

    // verify data integrity
    cy.get('body').should('contain', 'Notify the following:')
    cy.get('body').should('contain', 'Step #1:')
    cy.get('[data-testid=destination-chip]').should('contain', 'example.com')
    cy.get('[data-testid=destination-chip]').should(
      'not.contain',
      'webhook.site',
    )
  })
})
