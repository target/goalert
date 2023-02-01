import { Chance } from 'chance'
import { DateTime } from 'luxon'
import { pathPrefix, testScreen } from '../support/e2e'
const c = new Chance()

function testServices(screen: ScreenFormat): void {
  beforeEach(() => {
    window.localStorage.setItem('show_services_new_feature_popup', 'false')
  })
  describe('List Page', () => {
    let svc: Service
    beforeEach(() => {
      cy.createService()
        .then((s: Service) => {
          svc = s
        })
        .visit('/services')
    })

    it('should handle searching', () => {
      cy.get('ul[data-cy=paginated-list]').should('exist')
      // by name
      cy.pageSearch(svc.name)
      cy.get('body')
        .should('contain', svc.name)
        .should('contain', svc.description)
    })

    it('should handle searching with leading and trailing spaces', () => {
      const firstHalf = c.word({ length: 4 })
      const secondHalf = c.word({ length: 4 })
      cy.createService({ name: firstHalf + ' ' + secondHalf })
      cy.createService({ name: firstHalf + secondHalf })

      cy.get('ul[data-cy=paginated-list]').should('exist')
      // by name with spaces before and after
      // search is based on word-matching so spaces are irrelevant
      cy.pageSearch(' ' + svc.name + '  ')
      cy.get('body')
        .should('contain', svc.name)
        .should('contain', svc.description)

      // since front-end no longer trims spaces for search arguments, the literal search result for search string should show up, if it exists.
      cy.pageSearch(' ' + secondHalf)
      cy.get('body')
        .should('contain', firstHalf + ' ' + secondHalf)
        .should('not.contain', firstHalf + secondHalf)

      cy.pageSearch(firstHalf + secondHalf)
      cy.get('body')
        .should('contain', firstHalf + secondHalf)
        .should('not.contain', firstHalf + ' ' + secondHalf)
    })

    it('should link to details page', () => {
      cy.get('ul[data-cy=paginated-list]').should('exist')
      cy.pageSearch(svc.name)
      cy.get('#app').contains(svc.name).click()
      cy.url().should('eq', Cypress.config().baseUrl + `/services/${svc.id}`)
    })

    describe('Filtering', () => {
      let label1: Label
      let label2: Label // uses key/value from label1
      let intKey: IntegrationKey
      beforeEach(() => {
        cy.createLabel().then((l: Label) => {
          label1 = l

          cy.createLabel({
            key: label1.key, // same key, random value
          }).then((l: Label) => {
            label2 = l
          })
        })
        cy.createIntKey().then((i: IntegrationKey) => {
          intKey = i
        })
      })

      it('should open and close the filter popover', () => {
        // check that filter content doesn't exist yet
        cy.get('div[data-cy="label-key-container"]').should('not.exist')
        cy.get('div[data-cy="label-value-container"]').should('not.exist')
        cy.get('button[data-cy="filter-done"]').should('not.exist')
        cy.get('button[data-cy="filter-reset"]').should('not.exist')

        // open filter
        if (screen === 'mobile') {
          cy.get('[data-cy=app-bar] button[data-cy=open-search]').click()
        }
        cy.get('button[data-cy="services-filter-button"]').click()

        // check that filter content exists
        cy.get('div[data-cy="label-key-container"]').should('be.visible')
        cy.get('div[data-cy="label-value-container"]').should('be.visible')
        cy.get('button[data-cy="filter-done"]').should('be.visible')
        cy.get('button[data-cy="filter-reset"]').should('be.visible')

        // close filter
        cy.get('button[data-cy="filter-done"]').click()

        // check that filter content is removed from the dom
        cy.get('div[data-cy="label-key-container"]').should('not.exist')
        cy.get('div[data-cy="label-value-container"]').should('not.exist')
        cy.get('button[data-cy="filter-done"]').should('not.exist')
        cy.get('button[data-cy="filter-reset"]').should('not.exist')
      })

      it('should filter by label key', () => {
        // open filter
        if (screen === 'mobile') {
          cy.get('[data-cy=app-bar] button[data-cy=open-search]').click()
        }
        cy.get('button[data-cy="services-filter-button"]').click()

        cy.get('input[name="label-key"]').selectByLabel(label1.key)

        // close filter
        cy.get('button[data-cy="filter-done"]').click()

        cy.get('body')
          .should('contain', label1.svc.name)
          .should('contain', label1.svc.description)

        cy.get('body')
          .should('contain', label2.svc.name)
          .should('contain', label2.svc.description)
      })

      it('should not allow searching by label value with no key selected', () => {
        // open filter
        if (screen === 'mobile') {
          cy.get('[data-cy=app-bar] button[data-cy=open-search]').click()
        }
        cy.get('button[data-cy="services-filter-button"]').click()

        cy.get('input[name="label-value"]').should(
          'have.attr',
          'disabled',
          'disabled',
        )
      })

      it('should filter by label key and value', () => {
        // open filter
        if (screen === 'mobile') {
          cy.get('[data-cy=app-bar] button[data-cy=open-search]').click()
        }
        cy.get('button[data-cy="services-filter-button"]').click()

        cy.get('input[name="label-key"]').selectByLabel(label1.key)
        cy.get('input[name="label-value"]').selectByLabel(label1.value)

        // close filter
        cy.get('button[data-cy="filter-done"]').click()

        cy.get('body')
          .should('contain', label1.svc.name)
          .should('contain', label1.svc.description)

        // check that the second label with the same key but different value doesn't show
        cy.get('body')
          .should('not.contain', label2.svc.name)
          .should('not.contain', label2.svc.description)
      })

      it('should filter by integration key', () => {
        // open filter
        if (screen === 'mobile') {
          cy.get('[data-cy=app-bar] button[data-cy=open-search]').click()
        }
        cy.get('button[data-cy="services-filter-button"]').click()

        cy.get('input[name="integration-key"]').selectByLabel(intKey.id)

        // close filter
        cy.get('button[data-cy="filter-done"]').click()

        cy.get('body')
          .should('contain', intKey.svc.name)
          .should('contain', intKey.svc.description)
      })

      it('should reset label filters', () => {
        // open filter
        if (screen === 'mobile') {
          cy.get('[data-cy=app-bar] button[data-cy=open-search]').click()
        }
        cy.get('button[data-cy="services-filter-button"]').click()

        cy.get('input[name="label-key"]').selectByLabel(label1.key)
        cy.get('input[name="label-value"]').selectByLabel(label1.value)

        cy.get('input[name="label-key"]').should('have.value', label1.key)
        cy.get('input[name="label-value"]').should('have.value', label1.value)
      })

      it('should load in filter values from URL', () => {
        cy.visit('/services?search=' + label1.key + '=*')

        cy.get('body')
          .should('contain', label1.svc.name)
          .should('contain', label1.svc.description)

        cy.get('body')
          .should('contain', label2.svc.name)
          .should('contain', label2.svc.description)
      })
    })

    describe('Creation', () => {
      it('should allow canceling', () => {
        cy.pageFab()
        cy.dialogTitle('Create New Service')
        cy.dialogFinish('Cancel')
      })

      it(`should create a service when submitted`, () => {
        const name = 'SM Svc ' + c.word({ length: 8 })
        const description = c.word({ length: 10 })

        cy.pageFab()
        cy.dialogForm({ name, 'escalation-policy': svc.ep.name, description })
        cy.dialogFinish('Submit')

        // should be on details page
        cy.get('body').should('contain', name).should('contain', description)
      })

      it(`should create a service, with a generated EP, when submitted`, () => {
        const name = 'SM Svc ' + c.word({ length: 8 })
        const description = c.word({ length: 10 })

        cy.pageFab()
        cy.dialogForm({ name, description })
        cy.dialogFinish('Submit')

        // should be on details page
        cy.get('body').should('contain', name).should('contain', description)
      })
    })
  })

  describe('Details Page', () => {
    let svc: Service
    beforeEach(() =>
      cy.createService().then((s: Service) => {
        svc = s
        return cy.visit(`/services/${s.id}`)
      }),
    )

    it('should display correct information', () => {
      cy.get('body')
        .should('contain', svc.name)
        .should('contain', svc.description)
        .contains('a', svc.ep.name)
        .should(
          'have.attr',
          'href',
          pathPrefix() + `/escalation-policies/${svc.ep.id}`,
        )
    })

    it('should allow deleting the service', () => {
      cy.get('[data-cy="card-actions"]')
        .find('button[aria-label="Delete"]')
        .click()
      cy.dialogFinish('Confirm')
      cy.url().should('eq', Cypress.config().baseUrl + '/services')
      cy.pageSearch(svc.name)
      cy.get('body').should('contain', 'No results')
    })

    it('should handle updating information', () => {
      const name = 'SM Svc ' + c.word({ length: 8 })
      const description = c.word({ length: 10 })

      cy.createEP().then((ep: EP) => {
        cy.get('[data-cy="card-actions"]')
          .find('button[aria-label="Edit"]')
          .click()

        cy.dialogForm({ name, description, 'escalation-policy': ep.name })
        cy.dialogFinish('Submit')

        cy.get('body')
          .should('contain', name)
          .should('contain', description)
          .contains('a', ep.name)
          .should(
            'have.attr',
            'href',
            pathPrefix() + `/escalation-policies/${ep.id}`,
          )
      })
    })

    it('should navigate to and from metrics', () => {
      cy.navigateToAndFrom(
        screen,
        'Services',
        svc.name,
        'Metrics',
        `${svc.id}/alert-metrics`,
      )
    })

    it('should navigate to and from metrics', () => {
      cy.navigateToAndFrom(
        screen,
        'Services',
        svc.name,
        'Alerts',
        `${svc.id}/alerts`,
      )
    })

    it('should navigate to and from integration keys', () => {
      cy.navigateToAndFrom(
        screen,
        'Services',
        svc.name,
        'Integration Keys',
        `${svc.id}/integration-keys`,
      )
    })

    it('should navigate to and from heartbeat monitors', () => {
      cy.navigateToAndFrom(
        screen,
        'Services',
        svc.name,
        'Heartbeat Monitors',
        `${svc.id}/heartbeat-monitors`,
      )
    })

    it('should navigate to and from labels', () => {
      cy.navigateToAndFrom(
        screen,
        'Services',
        svc.name,
        'Labels',
        `${svc.id}/labels`,
      )
    })
  })

  describe('Heartbeat Monitors', () => {
    let monitor: HeartbeatMonitor
    beforeEach(() => {
      cy.createService().then((s: Service) =>
        cy
          .createHeartbeatMonitor({
            svcID: s.id,
            name: c.word({ length: 5 }) + ' Monitor',
            timeoutMinutes: Math.trunc(Math.random() * 10) + 5,
          })
          .then((m: HeartbeatMonitor) => {
            monitor = m
          })
          .visit(`/services/${s.id}/heartbeat-monitors`),
      )
    })

    it('should create a monitor', () => {
      const name = c.word({ length: 5 }) + ' Monitor'
      const timeoutMinutes = (Math.trunc(Math.random() * 10) + 5).toString()
      const invalidName = 'a'

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-monitor"]').click()
      }

      cy.dialogForm({ name: invalidName, timeoutMinutes })
      cy.dialogClick('Submit')
      cy.get('body').should('contain', 'Must be at least 2 characters')

      cy.dialogForm({ name, timeoutMinutes })
      cy.dialogFinish('Retry')

      cy.get('li').should('contain', name).should('contain', timeoutMinutes)
    })

    it('should edit a monitor', () => {
      const name = c.word({ length: 5 })
      const timeoutMinutes = (Math.trunc(Math.random() * 10) + 5).toString()

      cy.get('li')
        .should('contain', monitor.name)
        .find('div')
        .find('button[data-cy=other-actions]')
        .menu('Edit')

      cy.dialogForm({ name, timeoutMinutes })
      cy.dialogFinish('Submit')

      cy.get('li').should('contain', name)
      cy.get('li').should('contain', timeoutMinutes)
    })

    it('should delete a monitor', () => {
      cy.get('li')
        .should('contain', monitor.name)
        .find('div')
        .find('button[data-cy=other-actions]')
        .menu('Delete')

      cy.dialogFinish('Confirm')

      cy.get('li').should('not.contain', monitor.name)
      cy.get('li').should(
        'contain',
        'No heartbeat monitors exist for this service.',
      )
    })

    it('should handle canceling', () => {
      // cancel out of create
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-monitor"]').click()
      }
      cy.dialogTitle('Create New Heartbeat Monitor')
      cy.dialogFinish('Cancel')

      // cancel out of edit
      cy.get('li')
        .should('contain', monitor.name)
        .find('div')
        .find('button[data-cy=other-actions]')
        .menu('Edit')
      cy.dialogFinish('Cancel')

      // cancel out of delete
      cy.get('li')
        .should('contain', monitor.name)
        .find('div')
        .find('button[data-cy=other-actions]')
        .menu('Delete')
      cy.dialogFinish('Cancel')
    })
  })

  describe('Integration Keys', () => {
    let svc: Service
    beforeEach(() =>
      cy.createService().then((s: Service) => {
        svc = s
        return cy.visit(`/services/${svc.id}/integration-keys`)
      }),
    )

    const createKey = (type: string, name: string): void => {
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-key"]').click()
      }
      cy.dialogForm({ name, type })
      cy.dialogFinish('Submit')
    }

    it('should allow managing integration keys', () => {
      const name = 'SM Int ' + c.word({ length: 8 })

      cy.get('body').should('contain', 'No integration keys')
      ;['Generic API', 'Grafana'].forEach((type) => {
        createKey(type, name)

        cy.get('ul[data-cy=int-keys]').should('contain', name)

        // delete
        cy.get('ul[data-cy=int-keys')
          .contains('li', name)
          .find('button')
          .click()

        cy.dialogFinish('Confirm')
      })
    })

    it('should manage integration keys with mailgun disabled', () => {
      const domain = c.domain()
      const name = 'SM Int ' + c.word({ length: 8 })

      cy.get('body').should('contain', 'No integration keys')

      cy.updateConfig({
        Mailgun: {
          Enable: true,
          APIKey: 'key-' + c.string({ length: 32, pool: '0123456789abcdef' }),
          EmailDomain: domain,
        },
      })
      cy.reload()

      createKey('Email', name) // set email integration key
      cy.get('ul[data-cy=int-keys')
        .contains('li', name)
        .find('a')
        .should('have.attr', 'href')
        .and('include', 'mailto:')
        .and('include', domain)

      cy.updateConfig({ Mailgun: { Enable: false } })
      cy.reload()

      // check for disabled text
      cy.get('ul[data-cy=int-keys]').should(
        'contain',
        'Email integration keys are currently disabled.',
      )

      // check that dropdown type is hidden
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-key"]').click()
      }
      cy.get('input[name=type]').findByLabel('Email').should('not.exist')
    })
  })

  describe('Alerts', () => {
    let svc: Service
    beforeEach(() =>
      cy.createService().then((s: Service) => {
        svc = s
        return cy.visit(`/services/${s.id}/alerts`)
      }),
    )

    it('should create alert with prepopulated service', () => {
      const summary = c.sentence({ words: 3 })
      const details = c.word({ length: 10 })

      cy.pageFab()
      cy.dialogForm({ summary, details })
      cy.dialogClick('Next')
      cy.dialogContains(svc.name)
      cy.dialogClick('Submit')
      cy.dialogFinish('Done')
    })

    it('should allow ack/close all alerts', () => {
      cy.createAlert({ serviceID: svc.id })
      cy.createAlert({ serviceID: svc.id })
      cy.createAlert({ serviceID: svc.id })

      cy.reload()

      cy.get('ul[data-cy=paginated-list]').should('contain', 'UNACKNOWLEDGED')

      cy.get('button').contains('Acknowledge All').click()
      cy.dialogFinish('Confirm')

      cy.get('ul[data-cy=paginated-list]').should('contain', 'ACKNOWLEDGED')
      cy.get('ul[data-cy=paginated-list]').should(
        'not.contain',
        'UNACKNOWLEDGED',
      )

      cy.get('button').contains('Close All').click()
      cy.dialogFinish('Confirm')

      cy.get('body').should('contain', 'No results')
    })
  })

  describe('Labels', () => {
    let label: Label
    beforeEach(() =>
      cy.createLabel().then((l: Label) => {
        label = l
        return cy.visit(`/services/${l.svcID}/labels`)
      }),
    )

    it('should create a label', () => {
      const key = label.key
      const value = c.word({ length: 10 })

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-label"]').click()
      }
      cy.dialogForm({ key, value })
      cy.dialogFinish('Submit')
      cy.get('li').should('contain', key)
    })

    it('should set an existing label', () => {
      const key = label.key
      const value = c.word({ length: 10 })

      cy.createService().then((svc: Service) => {
        cy.visit(`/services/${svc.id}/labels`)
      })

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-label"]').click()
      }
      cy.dialogForm({ key, value })
      cy.dialogFinish('Submit')

      cy.get('li').should('contain', key)
      cy.get('li').should('contain', value)
    })

    it('should edit a label', () => {
      const key = label.key
      const value = c.word({ length: 10 })

      cy.get('li')
        .should('contain', key)
        .find('div')
        .find('button[data-cy=other-actions]')
        .menu('Edit')

      cy.dialogForm({ value })
      cy.dialogFinish('Submit')

      cy.get('li').should('contain', key)
      cy.get('li').should('contain', value)
    })

    it('should delete a label', () => {
      cy.get('li')
        .should('contain', label.key)
        .find('div')
        .find('button[data-cy=other-actions]')
        .menu('Delete')

      cy.dialogFinish('Confirm')

      cy.get('li').should('not.contain', label.key)
      cy.get('li').should('not.contain', label.value)
    })

    it('should search for a specific service with label', () => {
      cy.visit(`/services`)
      cy.get('ul[data-cy=paginated-list]').should('exist')
      cy.pageSearch(`${label.key}=${label.value}`)
      cy.get('body')
        .should('contain', label.svc.name)
        .should('contain', label.svc.description)
    })

    it('should search for a services without label', () => {
      cy.visit(`/services`)
      cy.get('ul[data-cy=paginated-list]').should('exist')
      cy.pageSearch(`${label.key}!=${label.value}`)
      cy.get('body')
        .should('not.contain', label.svc.name)
        .should('not.contain', label.svc.description)
    })

    it('should not be able to create a label when DisableLabelCreation is true', () => {
      const randomWord = c.word({
        length: 7,
      })
      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-label"]').click()
      }
      cy.get('input[name=key]').findByLabel(`Create "${randomWord}"`)

      cy.updateConfig({
        General: {
          DisableLabelCreation: true,
        },
      })
      cy.reload()

      if (screen === 'mobile') {
        cy.pageFab()
      } else {
        cy.get('button[data-testid="create-label"]').click()
      }
      cy.get('input[name=key]').type(`Create "${randomWord}"`)
      cy.get('[data-cy="select-dropdown"]').should('contain', 'No options')
    })
  })

  describe('Metrics', () => {
    let closedAlert: Alert
    let openAlert: Alert
    beforeEach(() =>
      cy
        .setTimeSpeed(0)
        .fastForward('-25h')
        .createAlert()
        .then((a: Alert) => {
          closedAlert = a
          cy.fastForward('1m')
          cy.ackAlert(a.id)
          cy.fastForward('1m')
          cy.closeAlert(a.id)
          cy.fastForward('25h')
          cy.setTimeSpeed(1) // resume the flow of time
          // non-closed alert
          return cy.createAlert({ serviceID: a.serviceID })
        })
        .then((a: Alert) => {
          openAlert = a
          return cy.visit(`/services/${a.serviceID}/alert-metrics`)
        }),
    )

    it('should display alert metrics', () => {
      const now = DateTime.local().minus({ day: 1 }).toLocaleString({
        month: 'short',
        day: 'numeric',
      })

      // summary doesn't load by default on mobile (until scrolled to)
      cy.get('[data-cy=metrics-table]')
        .should('contain', closedAlert.id)
        .should('not.contain', openAlert.id)

      cy.get('path[name="Alert Count"]')
        .should('have.length', 1)
        .trigger('mouseover')
      cy.get('[data-cy=metrics-count-graph]')
        .should('contain', now)
        .should('contain', 'Alert Count: 1')
        .should('contain', 'Escalated: 0') // no ep steps

      cy.get(`.recharts-line-dots circle[r=3]`).last().trigger('mouseover')
      cy.get('[data-cy=metrics-averages-graph]')
        .should('contain', now)
        .should('contain', 'Avg. Ack: 1 min')
        .should('contain', 'Avg. Close: 2 min')
    })
  })

  describe('Maintenance Mode', () => {
    let svc: Service
    let openAlert: Alert
    beforeEach(() => {
      cy.createUser().then((user: Profile) => {
        return cy.createService().then((s: Service) => {
          svc = s
          cy.createAlert({ serviceID: svc.id }).then((a: Alert) => {
            openAlert = a
          })
          cy.createEPStep({
            epID: s.epID,
            targets: [{ type: 'user', id: user.id }],
          }).then(() => s.id)
          return cy.visit(`/services/${s.id}`)
        })
      })
    })

    it('should start maintenance mode, display banners, and cancel', () => {
      cy.get('button[aria-label="Maintenance Mode"').click()
      cy.dialogFinish('Submit')

      cy.get('body').should('contain', 'Warning: In Maintenance Mode')
      cy.visit(`/services/${svc.id}/alerts`)
      cy.get('body').should('contain', 'Warning: In Maintenance Mode')
      cy.visit(`/services/${svc.id}/alerts/${openAlert.id}`)
      cy.get('body').should('contain', 'Warning: In Maintenance Mode')

      // verify escalate button is disabled
      cy.get('button[aria-label="Escalate disabled. In maintenance mode."]')
        .parent() // go 1 level up to focusable span
        .trigger('mouseover')
      cy.get('body').should(
        'contain',
        'Escalate disabled. In maintenance mode.',
      )

      // cancel maintenance mode
      cy.get('button[aria-label="Cancel Maintenance Mode"').click()
      cy.get('body').should('not.contain', 'Warning: In Maintenance Mode')
      cy.visit(`/services/${svc.id}/alerts`)
      cy.get('body').should('not.contain', 'Warning: In Maintenance Mode')
      cy.visit(`/services/${svc.id}`)
      cy.get('body').should('not.contain', 'Warning: In Maintenance Mode')
    })

    it('should not escalate to step 1 when alert created in maintenance mode', () => {
      cy.get('button[aria-label="Maintenance Mode"').click()
      cy.dialogFinish('Submit')

      const summary = 'test alert'
      cy.get('[data-cy=route-links] li').contains('Alerts').click()
      cy.get('button[aria-label="Create Alert"').click()
      cy.dialogForm({
        summary,
      })
      cy.dialogClick('Next')
      cy.dialogClick('Submit')
      cy.dialogFinish('Done')
      cy.get('p').contains(summary).click()
      cy.get('body').should('not.contain', 'Escalated to step #1')
    })
  })
}

testScreen('Services', testServices)
