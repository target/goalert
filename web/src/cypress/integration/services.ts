import { Chance } from 'chance'
import { testScreen } from '../support'
const c = new Chance()

function basePrefix(): string {
  const u = new URL(Cypress.config('baseUrl') as string)
  return u.pathname.replace(/\/$/, '')
}

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
      cy.get('ul[data-cy=apollo-list]').should('exist')
      // by name
      cy.pageSearch(svc.name)
      cy.get('body')
        .should('contain', svc.name)
        .should('contain', svc.description)
    })

    it('should handle searching with leading and trailing spaces', () => {
      cy.createService({ name: 'foobar' })
      cy.createService({ name: 'foo bar' })

      cy.get('ul[data-cy=apollo-list]').should('exist')
      // by name with spaces before and after
      // since search looks for literally the search string typed in, there would be no results for leading space + search string + 2 spaces
      cy.pageSearch(' ' + svc.name + '  ')
      cy.get('body')
        .should('not.contain', svc.name)
        .should('not.contain', svc.description)

      // since front-end no longer trims spaces for search arguments, the literal search result for search string should show up, if it exists.
      cy.pageSearch(' bar')
      cy.get('body')
        .should('contain', 'foo bar')
        .should('not.contain', 'foobar')

      cy.pageSearch('foobar')
      cy.get('body')
        .should('contain', 'foobar')
        .should('not.contain', 'foo bar')
    })

    it('should link to details page', () => {
      cy.get('ul[data-cy=apollo-list]').should('exist')
      cy.pageSearch(svc.name)
      cy.get('#app').contains(svc.name).click()
      cy.url().should('eq', Cypress.config().baseUrl + `/services/${svc.id}`)
    })

    describe('Filtering', () => {
      let label1: Label
      let label2: Label // uses key/value from label1
      beforeEach(() => {
        cy.createLabel().then((l: Label) => {
          label1 = l

          cy.createLabel({
            key: label1.key, // same key, random value
          }).then((l: Label) => {
            label2 = l
          })
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

        cy.get('div[name="label-value"]').should(
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
          basePrefix() + `/escalation-policies/${svc.ep.id}`,
        )
    })

    it('should allow deleting the service', () => {
      cy.pageAction('Delete')
      cy.dialogFinish('Confirm')
      cy.url().should('eq', Cypress.config().baseUrl + '/services')
      cy.pageSearch(svc.name)
      cy.get('body').should('contain', 'No results')
    })

    it('should handle updating information', () => {
      const name = 'SM Svc ' + c.word({ length: 8 })
      const description = c.word({ length: 10 })

      cy.createEP().then((ep: EP) => {
        cy.pageAction('Edit')

        cy.dialogForm({ name, description, 'escalation-policy': ep.name })
        cy.dialogFinish('Submit')

        cy.get('body')
          .should('contain', name)
          .should('contain', description)
          .contains('a', ep.name)
          .should(
            'have.attr',
            'href',
            basePrefix() + `/escalation-policies/${ep.id}`,
          )
      })
    })

    it('should navigate to and from alerts', () => {
      cy.navigateToAndFrom(
        screen,
        'Service Details',
        svc.name,
        'Alerts',
        `${svc.id}/alerts`,
      )
    })

    it('should navigate to and from integration keys', () => {
      cy.navigateToAndFrom(
        screen,
        'Service Details',
        svc.name,
        'Integration Keys',
        `${svc.id}/integration-keys`,
      )
    })

    it('should navigate to and from heartbeat monitors', () => {
      cy.navigateToAndFrom(
        screen,
        'Service Details',
        svc.name,
        'Heartbeat Monitors',
        `${svc.id}/heartbeat-monitors`,
      )
    })

    it('should navigate to and from labels', () => {
      cy.navigateToAndFrom(
        screen,
        'Service Details',
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

      cy.pageFab()

      cy.dialogForm({ name, timeoutMinutes })
      cy.dialogFinish('Submit')

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
      cy.pageFab()
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
      cy.pageFab()
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
      cy.pageFab()
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

      cy.get('ul[data-cy=apollo-list]').should('contain', 'UNACKNOWLEDGED')

      cy.pageAction('Acknowledge All')
      cy.dialogFinish('Confirm')

      cy.get('ul[data-cy=apollo-list]').should('contain', 'ACKNOWLEDGED')
      cy.get('ul[data-cy=apollo-list]').should('not.contain', 'UNACKNOWLEDGED')

      cy.pageAction('Close All')
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

      cy.pageFab()
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

      cy.pageFab()
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
      cy.get('ul[data-cy=apollo-list]').should('exist')
      cy.pageSearch(`${label.key}=${label.value}`)
      cy.get('body')
        .should('contain', label.svc.name)
        .should('contain', label.svc.description)
    })

    it('should search for a services without label', () => {
      cy.visit(`/services`)
      cy.get('ul[data-cy=apollo-list]').should('exist')
      cy.pageSearch(`${label.key}!=${label.value}`)
      cy.get('body')
        .should('not.contain', label.svc.name)
        .should('not.contain', label.svc.description)
    })

    it('should not be able to create a label when DisableLabelCreation is true', () => {
      const randomWord = c.word({ length: 7 })
      cy.pageFab()
      cy.get('input[name=key]').findByLabel(`Create "${randomWord}"`)

      cy.updateConfig({
        General: {
          DisableLabelCreation: true,
        },
      })
      cy.reload()

      cy.pageFab()
      cy.get('input[name=key]')
        .findByLabel(`Create "${randomWord}"`)
        .should('not.exist')
    })
  })
}

testScreen('Services', testServices)
