import { Chance } from 'chance'

import { testScreen } from '../support'
const c = new Chance()

testScreen('Alerts', testAlerts)

function testAlerts(screen: ScreenFormat) {
  describe('Alerts List', () => {
    let alert: Alert
    beforeEach(() => {
      cy.createAlert()
        .then(a => {
          alert = a
        })
        .visit('/alerts?allServices=1')
    })

    it('should handle searching by number', () => {
      // by id
      cy.pageSearch(alert.number.toString())
      cy.get('body')
        .should('contain', alert.summary)
        .should('contain', alert.number)
        .should('contain', alert.service.name)
      cy.get('ul[data-cy=alerts-list] li').should('have.length', 1)
    })
    it('should handle searching by summary', () => {
      // by summary
      cy.pageSearch(alert.summary)
      cy.get('body')
        .should('contain', alert.summary)
        .should('contain', alert.number)
        .should('contain', alert.service.name)
      cy.get('ul[data-cy=alerts-list] li').should('have.length', 1)
    })

    it('should handle searching by service name', () => {
      // by service
      cy.pageSearch(alert.service.name)
      cy.get('body')
        .should('contain', alert.summary)
        .should('contain', alert.number)
        .should('contain', alert.service.name)
      cy.get('ul[data-cy=alerts-list] li').should('have.length', 1)
    })

    it('should handle toggling show by favorites filter', () => {
      cy.visit('/alerts')
      cy.get('body').should('contain', 'No results') // mock data has no favorited services-- 0 alerts should show
      cy.get('button[aria-label="Filter Alerts"]').click()
      cy.get('span[data-cy=toggle-favorites]').click() // set to false (see all alerts)
      cy.get('body').should('not.contain', 'No results') // mock alerts should show again
    })

    it('should load more list items when scrolling to the bottom', () => {
      const summary = c.word()

      cy.createManyAlerts(50, { summary }).then(() => {
        cy.visit('/alerts?allServices=1&filter=all&search=' + summary)
        cy.get('[data-cy=alerts-list] li').should('have.length', 25)
        cy.get('[id="content"]').scrollTo('bottom')
        cy.get('[data-cy=alerts-list] li').should('have.length', 50)
      })
    })

    describe('Item', () => {
      beforeEach(() => cy.pageSearch(alert.number.toString()))
      it('should link to the details page', () => {
        cy.get('ul[data-cy=alerts-list]')
          .contains(alert.number.toString())
          .click()

        cy.url().should(
          'eq',
          Cypress.config().baseUrl + `/alerts/${alert.number}`,
        )
      })
    })
  })

  describe('Alerts checkboxes', () => {
    let svc: Service
    let alert1: Alert
    let alert2: Alert
    let alert3: Alert

    beforeEach(() => {
      cy.createService({ ep: { stepCount: 1 } }).then(s => {
        svc = s
        cy.createAlert({ serviceID: svc.id }).then(a => {
          alert1 = a
        })
        cy.createAlert({ serviceID: svc.id }).then(a => {
          alert2 = a
        })
        cy.createAlert({ serviceID: svc.id }).then(a => {
          alert3 = a
        })

        cy.visit(`/alerts?allServices=1&search=${svc.name}`)

        // wait for list to fully load before beginning tests
        return cy
          .get('[data-cy=alerts-list] [role=button]')
          .should('have.length', 3)
      })
    })

    it('should select and deselect all alerts from the header checkbox', () => {
      cy.get('span[data-cy=select-all] input').check()

      cy.get(`span[data-cy=alert-${alert1.number}] input`).should('be.checked')
      cy.get(`span[data-cy=alert-${alert2.number}] input`).should('be.checked')
      cy.get(`span[data-cy=alert-${alert3.number}] input`).should('be.checked')

      cy.get('span[data-cy=select-all] input').uncheck()

      cy.get(`span[data-cy=alert-${alert1.number}] input`).should(
        'not.be.checked',
      )
      cy.get(`span[data-cy=alert-${alert2.number}] input`).should(
        'not.be.checked',
      )
      cy.get(`span[data-cy=alert-${alert3.number}] input`).should(
        'not.be.checked',
      )
    })

    it('should select some alerts and deselect all from the header checkbox', () => {
      cy.get(`span[data-cy=alert-${alert1.number}] input`).check()
      cy.get(`span[data-cy=alert-${alert2.number}] input`).check()

      cy.get('span[data-cy=select-all] input').click()

      cy.get(`span[data-cy=alert-${alert1.number}] input`).should(
        'not.be.checked',
      )
      cy.get(`span[data-cy=alert-${alert2.number}] input`).should(
        'not.be.checked',
      )
      cy.get(`span[data-cy=alert-${alert2.number}] input`).should(
        'not.be.checked',
      )
    })

    it('should select and deselect all alerts from the header checkbox menu', () => {
      cy.get('[data-cy=checkboxes-menu] [data-cy=other-actions]').menu('All')

      cy.get(`span[data-cy=alert-${alert1.number}] input`).should('be.checked')
      cy.get(`span[data-cy=alert-${alert2.number}] input`).should('be.checked')
      cy.get(`span[data-cy=alert-${alert3.number}] input`).should('be.checked')

      cy.get('[data-cy=checkboxes-menu] [data-cy=other-actions]').menu('None')
      cy.get(`span[data-cy=alert-${alert1.number}]`).should('not.be.checked')
      cy.get(`span[data-cy=alert-${alert2.number}]`).should('not.be.checked')
      cy.get(`span[data-cy=alert-${alert3.number}]`).should('not.be.checked')
    })

    it('should acknowledge, escalate, and close multiple alerts', () => {
      cy.get('span[data-cy=select-all] input')
        .should('not.be.checked')
        .click()

      cy.get('button[data-cy=acknowledge]').click()

      cy.get('ul[data-cy=alerts-list]').should('not.contain', 'UNACKNOWLEDGED')

      cy.get('span[data-cy=select-all] input')
        .should('not.be.checked')
        .click()

      cy.get('button[data-cy=escalate]').click()
      cy.get('ul[data-cy=alerts-list]').should('contain', 'UNACKNOWLEDGED')

      cy.get('span[data-cy=select-all] input')
        .should('not.be.checked')
        .click()

      cy.get('button[data-cy=close]').click()
      cy.get('ul[data-cy=alerts-list-no-data]').should('contain', 'No results')
    })

    it('should update some alerts', () => {
      // prep
      cy.get(`span[data-cy=alert-${alert1.number}] input`).check()
      cy.get('button[data-cy=acknowledge]').click()
      cy.get('button[data-cy=acknowledge]').should('not.exist')

      cy.get(`[data-cy=alert-${alert1.number}]`)
        .parent('[role=button]')
        .should('not.contain', 'UNACKNOWLEDGED')
      cy.get(`[data-cy=alert-${alert2.number}]`)
        .parent('[role=button]')
        .should('contain', 'UNACKNOWLEDGED')
      cy.get(`[data-cy=alert-${alert3.number}]`)
        .parent('[role=button]')
        .should('contain', 'UNACKNOWLEDGED')

      cy.get(`[data-cy=select-all] input`).check()

      cy.get('button[data-cy=acknowledge]').click()
      cy.get('span[data-cy=update-message]').should(
        'contain',
        '2 of 3 alerts updated',
      )
      cy.get(`[data-cy=alert-${alert1.number}]`)
        .parent('[role=button]')
        .should('not.contain', 'UNACKNOWLEDGED')
      cy.get(`[data-cy=alert-${alert2.number}]`)
        .parent('[role=button]')
        .should('not.contain', 'UNACKNOWLEDGED')
      cy.get(`[data-cy=alert-${alert3.number}]`)
        .parent('[role=button]')
        .should('not.contain', 'UNACKNOWLEDGED')
    })

    it('should NOT acknowledge acknowledged alerts', () => {
      // ack first two
      cy.get(`span[data-cy=alert-${alert1.number}] input`).check()
      cy.get(`span[data-cy=alert-${alert2.number}] input`).check()
      cy.get('button[data-cy=acknowledge]').click()

      // ack
      // ack
      // unack
      cy.get(`[data-cy=alert-${alert1.number}]`)
        .parent('[role=button]')
        .should('not.contain', 'UNACKNOWLEDGED')
      cy.get(`[data-cy=alert-${alert2.number}]`)
        .parent('[role=button]')
        .should('not.contain', 'UNACKNOWLEDGED')
      cy.get(`[data-cy=alert-${alert3.number}]`)
        .parent('[role=button]')
        .should('contain', 'UNACKNOWLEDGED')

      // ack first two again (noop)
      cy.get(`span[data-cy=alert-${alert1.number}] input`).check()
      cy.get(`span[data-cy=alert-${alert2.number}] input`).check()
      cy.get('button[data-cy=acknowledge]').click()

      cy.get('span[data-cy=update-message]').should(
        'contain',
        '0 of 2 alerts updated',
      )

      // ack all three
      cy.get(`span[data-cy=alert-${alert1.number}] input`).check()
      cy.get(`span[data-cy=alert-${alert2.number}] input`).check()
      cy.get(`span[data-cy=alert-${alert3.number}] input`).check()
      cy.get('button[data-cy=acknowledge]').click()

      // first two already acked, third now acked
      cy.get('span[data-cy=update-message]').should(
        'contain',
        '1 of 3 alerts updated',
      )
    })
  })

  describe('Alert Creation', () => {
    beforeEach(() => cy.visit('/alerts?allServices=1'))

    let svc1: Service
    let svc2: Service

    beforeEach(() => {
      cy.createService().then(s => {
        svc1 = s
      })

      cy.createService().then(s => {
        svc2 = s
      })
    })

    it('should create an alert for two services', () => {
      const summary = c.sentence({
        words: 3,
      })
      const details = c.word({ length: 10 })

      cy.pageFab()

      // Alert Info
      cy.dialogTitle('Create New Alert')
      cy.dialogForm({
        summary,
        details,
      })
      cy.dialogClick('Next')

      // Service Selection
      cy.dialogForm({ serviceSearch: svc1.name })
      cy.get('ul')
        .contains(svc1.name)
        .click()
      cy.dialogForm({ serviceSearch: svc2.name })
      cy.get('ul')
        .contains(svc2.name)
        .click()
      cy.dialogForm({ serviceSearch: '' })

      cy.dialogContains('Selected Services (2)')
      cy.dialogClick('Next')

      // Confirm
      cy.dialogContains(svc1.name)
      cy.dialogContains(svc2.name)
      cy.get('[data-cy=service-chip]').contains(svc1.name)
      cy.get('[data-cy=service-chip]').contains(svc2.name)
      cy.dialogClick('Submit')

      // Review
      cy.dialogContains('Successfully created 2 alerts')
      cy.dialogFinish('Done')
    })
  })

  describe('Alert Details', () => {
    let alert: Alert
    beforeEach(() => {
      cy.createAlert({ service: { ep: { stepCount: 1 } } }).then(a => {
        alert = a
        return cy.visit(`/alerts/${a.number}`)
      })
    })

    if (screen === 'widescreen') {
      it('should link to the escalation policy', () => {
        cy.get('body')
          .contains('a', 'Escalation Policy')
          .click()
          .url()
          .should(
            'eq',
            Cypress.config().baseUrl +
              `/escalation-policies/${alert.service.ep.id}`,
          )
      })
    }

    it('should link to the service', () => {
      cy.get('body')
        .contains('a', alert.service.name)
        .click()
        .url()
        .should(
          'eq',
          Cypress.config().baseUrl + `/services/${alert.service.id}`,
        )
    })

    it('should have proper data', () => {
      cy.get('body').should('contain', alert.details)
      cy.get('body').should('contain', alert.summary)
      cy.get('body').should('contain', 'Created by Cypress User')
      cy.get('body').should('contain', 'UNACKNOWLEDGED')
    })

    it('should allow the user to take action', () => {
      // ack
      cy.pageAction('Acknowledge')

      cy.get('body').should('contain', 'ACKNOWLEDGED')
      cy.get('body').should('not.contain', 'UNACKNOWLEDGED')
      cy.get('body').should('contain', 'Acknowledged by Cypress User')

      // Escalation
      cy.pageAction('Escalate')
      cy.get('body').should('contain', 'Escalation requested by Cypress User')

      // close
      cy.pageAction('Close')
      cy.get('body').should('contain', 'Closed by Cypress User')
      cy.get('body').should('contain', 'CLOSED')
    })
  })
  describe('Alert Details Logs', () => {
    let logs: AlertLogs
    beforeEach(() => {
      cy.createAlertLogs({ count: 200 }).then(_logs => {
        logs = _logs
        return cy.visit(`/alerts/${logs.alert.number}`)
      })
    })

    it('should see load more, click, and no longer see load more', () => {
      cy.get('ul[data-cy=alert-logs] li').should('have.length', 35)
      cy.get('body').should('contain', 'Load More')
      cy.get('[data-cy=load-more-logs]').click()
      cy.get('ul[data-cy=alert-logs] li').should('have.length', 184)
      cy.get('body').should('contain', 'Load More')
      cy.get('[data-cy=load-more-logs]').click()

      // create plus any engine events should be 200+
      cy.get('ul[data-cy=alert-logs] li').should('have.length.gt', 200)
      cy.get('body').should('not.contain', 'Load More')
    })
  })
}
