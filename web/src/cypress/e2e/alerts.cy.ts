import { Chance } from 'chance'

import { pathPrefix, testScreen } from '../support/e2e'
const c = new Chance()

function testAlerts(screen: ScreenFormat): void {
  describe('Alerts List', () => {
    let alert: Alert
    beforeEach(() => {
      cy.createAlert()
        .then((a: Alert) => {
          alert = a
        })
        .visit('/alerts?allServices=1')
    })

    it('should handle searching by id', () => {
      // by id
      cy.pageSearch(alert.id.toString())
      cy.get('body')
        .should('contain', alert.summary)
        .should('contain', alert.id)
        .should('contain', alert.service.name)
      cy.get('ul[data-cy=paginated-list] li a').should('have.length', 1)
    })

    it('should handle searching by summary', () => {
      // by summary
      cy.pageSearch(alert.summary)
      cy.get('body')
        .should('contain', alert.summary)
        .should('contain', alert.id)
        .should('contain', alert.service.name)
      cy.get('ul[data-cy=paginated-list] li a').should('have.length', 1)
    })

    it('should handle searching by service name', () => {
      // by service
      cy.pageSearch(alert.service.name)
      cy.get('body')
        .should('contain', alert.summary)
        .should('contain', alert.id)
        .should('contain', alert.service.name)
      cy.get('ul[data-cy=paginated-list] li a').should('have.length', 1)
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
        cy.get('[data-cy=paginated-list] li').should('contain', summary)
        cy.get('[data-cy=paginated-list] li a').should('have.length', 25)
        cy.get('[id="content"]').scrollTo('bottom')
        cy.get('[data-cy=paginated-list] li a').should('have.length', 50)
      })
    })

    describe('Item', () => {
      beforeEach(() => cy.pageSearch(alert.id.toString()))
      it('should link to the details page', () => {
        cy.get('ul[data-cy=paginated-list] li a').should('have.lengthOf', 1)
        cy.get('ul[data-cy=paginated-list]')
          .contains(alert.id.toString())
          .click()

        cy.url().should(
          'eq',
          Cypress.config().baseUrl +
            `/services/${alert.serviceID}/alerts/${alert.id}`,
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
      cy.createService({ ep: { stepCount: 1 } }).then((s: Service) => {
        svc = s
        cy.createAlert({ serviceID: svc.id }).then((a: Alert) => {
          alert1 = a
        })
        cy.createAlert({ serviceID: svc.id }).then((a: Alert) => {
          alert2 = a
        })
        cy.createAlert({ serviceID: svc.id }).then((a: Alert) => {
          alert3 = a
        })

        cy.visit(`/alerts?allServices=1&search=${svc.name}`)

        // wait for list to fully load before beginning tests
        return cy
          .get('[data-cy=paginated-list] [type=checkbox]')
          .should('have.length', 3)
      })
    })

    it('should select and deselect all alerts from the header checkbox', () => {
      cy.get('span[data-cy=select-all] input').check()

      cy.get(`span[data-cy=item-${alert1.id}] input`).should('be.checked')
      cy.get(`span[data-cy=item-${alert2.id}] input`).should('be.checked')
      cy.get(`span[data-cy=item-${alert3.id}] input`).should('be.checked')

      cy.get('span[data-cy=select-all] input').uncheck()

      cy.get(`span[data-cy=item-${alert1.id}] input`).should('not.be.checked')
      cy.get(`span[data-cy=item-${alert2.id}] input`).should('not.be.checked')
      cy.get(`span[data-cy=item-${alert3.id}] input`).should('not.be.checked')
    })

    it('should select some alerts and deselect all from the header checkbox', () => {
      cy.get(`span[data-cy=item-${alert1.id}] input`).check()
      cy.get(`span[data-cy=item-${alert2.id}] input`).check()

      cy.get('span[data-cy=select-all] input').click()

      cy.get(`span[data-cy=item-${alert1.id}] input`).should('not.be.checked')
      cy.get(`span[data-cy=item-${alert2.id}] input`).should('not.be.checked')
      cy.get(`span[data-cy=item-${alert2.id}] input`).should('not.be.checked')
    })

    it('should select and deselect all alerts from the header checkbox menu', () => {
      cy.get('[data-cy=checkboxes-menu] [data-cy=other-actions]').menu('All')

      cy.get(`span[data-cy=item-${alert1.id}] input`).should('be.checked')
      cy.get(`span[data-cy=item-${alert2.id}] input`).should('be.checked')
      cy.get(`span[data-cy=item-${alert3.id}] input`).should('be.checked')

      cy.get('[data-cy=checkboxes-menu] [data-cy=other-actions]').menu('None')
      cy.get(`span[data-cy=item-${alert1.id}]`).should('not.be.checked')
      cy.get(`span[data-cy=item-${alert2.id}]`).should('not.be.checked')
      cy.get(`span[data-cy=item-${alert3.id}]`).should('not.be.checked')
    })

    it('should acknowledge, escalate, and close multiple alerts', () => {
      cy.get('span[data-cy=select-all] input').should('not.be.checked').click()

      cy.get('button[aria-label=Acknowledge]').click()

      cy.get('ul[data-cy=paginated-list] li a').should(
        'have.length.at.least',
        1,
      )
      cy.get('ul[data-cy=paginated-list] li a').should(
        'not.contain',
        'UNACKNOWLEDGED',
      )

      cy.get('[data-cy=paginated-list] input')
        .first()
        .should('not.be.checked')
        .click()

      cy.get('button[aria-label=Escalate]').click()
      cy.get('ul[data-cy=paginated-list] li a').should(
        'have.length.at.least',
        1,
      )
      cy.get('ul[data-cy=paginated-list] li a').should(
        'contain',
        'UNACKNOWLEDGED',
      )

      cy.get('span[data-cy=select-all] input').should('not.be.checked').click()

      cy.get('button[aria-label=Close]').should('have.length', 1).click()
      cy.get('ul[data-cy=paginated-list]').should('contain', 'No results')
    })

    it('should update some alerts', () => {
      // prep
      cy.get(`span[data-cy=item-${alert1.id}] input`).check()
      cy.get('button[aria-label=Acknowledge]').click()
      cy.get('button[aria-label=Acknowledge]').should('not.exist')

      cy.get(
        `[href="${pathPrefix()}/services/${alert1.serviceID}/alerts/${
          alert1.id
        }"]`,
      ).should('not.contain', 'UNACKNOWLEDGED')
      cy.get(
        `[href="${pathPrefix()}/services/${alert2.serviceID}/alerts/${
          alert2.id
        }"]`,
      ).should('contain', 'UNACKNOWLEDGED')
      cy.get(
        `[href="${pathPrefix()}/services/${alert3.serviceID}/alerts/${
          alert3.id
        }"]`,
      ).should('contain', 'UNACKNOWLEDGED')

      cy.reload()
      cy.get(`span[data-cy=item-${alert1.id}] input`).check()
      cy.get(`span[data-cy=item-${alert2.id}] input`).check()
      cy.get(`span[data-cy=item-${alert3.id}] input`).check()

      cy.get('button[aria-label=Acknowledge]').click()
      cy.get('[role="alert"]').should('contain', '2 of 3 alerts updated')
      cy.get(
        `[href="${pathPrefix()}/services/${alert1.serviceID}/alerts/${
          alert1.id
        }"]`,
      ).should('not.contain', 'UNACKNOWLEDGED')
      cy.get(
        `[href="${pathPrefix()}/services/${alert2.serviceID}/alerts/${
          alert2.id
        }"]`,
      ).should('not.contain', 'UNACKNOWLEDGED')
      cy.get(
        `[href="${pathPrefix()}/services/${alert3.serviceID}/alerts/${
          alert3.id
        }"]`,
      ).should('not.contain', 'UNACKNOWLEDGED')
    })

    it('should not acknowledge acknowledged alerts', () => {
      const prefix = new URL(Cypress.config().baseUrl || '').pathname.replace(
        /\/$/,
        '',
      )
      // ack first two
      cy.get(`span[data-cy=item-${alert1.id}] input`).check()
      cy.get(`span[data-cy=item-${alert2.id}] input`).check()
      cy.get('button[aria-label=Acknowledge]').click()

      // ack
      // ack
      // unack
      cy.get(
        `[href="${prefix}/services/${alert1.serviceID}/alerts/${alert1.id}"]`,
      ).should('not.contain', 'UNACKNOWLEDGED')
      cy.get(
        `[href="${prefix}/services/${alert2.serviceID}/alerts/${alert2.id}"]`,
      ).should('not.contain', 'UNACKNOWLEDGED')
      cy.get(
        `[href="${prefix}/services/${alert3.serviceID}/alerts/${alert3.id}"]`,
      ).should('contain', 'UNACKNOWLEDGED')

      // ack first two again (noop)
      cy.reload()
      cy.get(`span[data-cy=item-${alert1.id}] input`).check()
      cy.get(`span[data-cy=item-${alert2.id}] input`).check()
      cy.get('button[aria-label=Acknowledge]').click()

      cy.get('[role="alert"]').should('contain', '0 of 2 alerts updated')

      // ack all three
      cy.reload()
      cy.get(`span[data-cy=item-${alert1.id}] input`).check()
      cy.get(`span[data-cy=item-${alert2.id}] input`).check()
      cy.get(`span[data-cy=item-${alert3.id}] input`).check()
      cy.get('button[aria-label=Acknowledge]').click()

      // first two already acked, third now acked
      cy.get('[role="alert"]').should('contain', '1 of 3 alerts updated')
    })
  })

  describe('Alert Creation', () => {
    beforeEach(() => cy.visit('/alerts?allServices=1'))

    let svc1: Service
    let svc2: Service

    beforeEach(() => {
      cy.createService().then((s: Service) => {
        svc1 = s
      })

      cy.createService().then((s: Service) => {
        svc2 = s
      })
    })

    it('should create an alert for two services', () => {
      const summary = c.sentence({
        words: 3,
      })
      const details = c.word({ length: 10 })

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button').contains('Create Alert').click()
      }

      // Alert Info
      cy.dialogTitle('Create New Alert')
      cy.dialogForm({
        summary,
        details,
      })
      cy.dialogClick('Next')

      // Service Selection
      cy.dialogForm({ serviceSearch: svc1.name })
      cy.get('ul').contains(svc1.name).click()
      cy.dialogForm({ serviceSearch: svc2.name })
      cy.get('ul').contains(svc2.name).click()
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

  it('should redirect to the correct service', () => {
    cy.createAlert().then((a: Alert) => {
      cy.visit(`/services/bobs-service/alerts/${a.id}`)
      cy.url().should(
        'eq',
        Cypress.config().baseUrl + `/services/${a.serviceID}/alerts/${a.id}`,
      )
    })
  })

  describe('Alert Details', () => {
    let alert: Alert
    beforeEach(() => {
      cy.createAlert({ service: { ep: { stepCount: 1 } } }).then((a: Alert) => {
        alert = a
        return cy.visit(`/alerts/${a.id}`)
      })
    })

    if (screen === 'widescreen') {
      it('should link to the escalation policy', () => {
        cy.get('body').contains('a', 'Escalation Policy').click()
        cy.url().should(
          'eq',
          Cypress.config().baseUrl +
            `/escalation-policies/${alert.service.ep.id}`,
        )
      })
    }

    it('should link to the service', () => {
      cy.get('body').contains('a', alert.service.name).click()
      cy.url().should(
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
      cy.get('button').contains('Acknowledge').click()

      cy.get('body').should('contain', 'ACKNOWLEDGED')
      cy.get('body').should('not.contain', 'UNACKNOWLEDGED')
      cy.get('body').should('contain', 'Acknowledged by Cypress User')

      // escalate
      cy.get('button').contains('Escalate').click()
      cy.get('body').should('contain', 'Escalation requested by Cypress User')
      cy.reload() // allows time for escalation request to process

      // close
      cy.get('button').contains('Close').click()
      cy.get('body').should('contain', 'Closed by Cypress User')
      cy.get('body').should('contain', 'CLOSED')
    })

    it('should set alert notes', () => {
      // set all notes, checking carefully because of async setState
      cy.get('body').should('contain.text', 'Is this alert noise?')
      cy.get('[data-cy="False positive"] input[type="checkbox"]').check()
      cy.get('[data-cy="False positive"] input[type="checkbox"]').should(
        'be.checked',
      )
      cy.get('[data-cy="Not actionable"] input[type="checkbox"]').check()
      cy.get('[data-cy="Not actionable"] input[type="checkbox"]').should(
        'be.checked',
      )
      cy.get('[data-cy="Poor details"] input[type="checkbox"]').check()
      cy.get('[data-cy="Poor details"] input[type="checkbox"]').should(
        'be.checked',
      )
      cy.get('[placeholder="Other (please specify)"]').type('Test')

      // submit
      cy.get('button[aria-label="Submit alert notes"]').should(
        'not.be.disabled',
      )
      cy.get('button[aria-label="Submit alert notes"]').click()
      cy.get('label').contains('False positive').should('not.exist')

      // see notice
      const noticeTitle = 'Info: This alert has been marked as noise'
      cy.get('body').should('contain.text', noticeTitle)
      cy.get('body').should(
        'contain.text',
        'Reasons: False positive, Not actionable, Poor details, Test',
      )

      // undo
      cy.get('button[aria-label="Reset alert notes"]').click()
      cy.get('body').should('not.contain.text', noticeTitle)
      cy.get('[data-cy="False positive"] input[type="checkbox"]').should(
        'not.be.checked',
      )
    })
  })

  describe('Alert Details Logs', () => {
    let logs: AlertLogs
    beforeEach(() => {
      cy.createAlertLogs({ count: 200 }).then((_logs: AlertLogs) => {
        logs = _logs
        return cy.visit(`/alerts/${logs.alert.id}`)
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

testScreen('Alerts', testAlerts)
