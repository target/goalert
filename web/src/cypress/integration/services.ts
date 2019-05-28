import { Chance } from 'chance'
import { testScreen } from '../support'
const c = new Chance()

testScreen('Services', testServices)

function testServices(screen: ScreenFormat) {
  beforeEach(() => {
    window.localStorage.setItem('show_services_new_feature_popup', 'false')
  })
  describe('List Page', () => {
    let svc: Service
    beforeEach(() => {
      cy.createService()
        .then(s => {
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
      cy.get('ul[data-cy=apollo-list]').should('exist')
      // by name with spaces before and after
      cy.pageSearch(' ' + svc.name + '  ')
      cy.get('body')
        .should('contain', svc.name)
        .should('contain', svc.description)
    })

    it('should link to details page', () => {
      cy.get('ul[data-cy=apollo-list]').should('exist')
      cy.pageSearch(svc.name)
      cy.get('#app')
        .contains(svc.name)
        .click()
      cy.location('pathname').should('eq', `/services/${svc.id}`)
    })

    describe('Creation', () => {
      it('should allow canceling', () => {
        cy.pageFab()
        cy.get('div[role=dialog]').should('contain', 'Create New Service')
        cy.get('div[role=dialog]')
          .contains('button', 'Cancel')
          .click()
        cy.get('div[role=dialog]').should('not.exist')
      })

      it(`should create a service when submitted`, () => {
        cy.pageFab()

        cy.get('div[role=dialog]').as('dialog')

        const name = 'SM Svc ' + c.word({ length: 8 })
        const description = c.word({ length: 10 })

        cy.get('@dialog')
          .find('input[name=name]')
          .type(name)
        cy.get('@dialog')
          .find('input[name=escalation-policy]')
          .selectByLabel(svc.ep.name)
        cy.get('@dialog')
          .find('textarea[name=description]')
          .type(description)

        cy.get('@dialog')
          .contains('button', 'Submit')
          .click()

        // should be on details page
        cy.get('body')
          .should('contain', name)
          .should('contain', description)
      })

      it(`should create a service, with a generated EP, when submitted`, () => {
        cy.pageFab()

        cy.get('div[role=dialog]').as('dialog')

        const name = 'SM Svc ' + c.word({ length: 8 })
        const description = c.word({ length: 10 })

        cy.get('@dialog')
          .find('input[name=name]')
          .type(name)
        cy.get('@dialog')
          .find('textarea[name=description]')
          .type(description)

        cy.get('@dialog')
          .contains('button', 'Submit')
          .click()

        // should be on details page
        cy.get('body')
          .should('contain', name)
          .should('contain', description)
      })
    })
  })

  describe('Details Page', () => {
    let svc: Service
    beforeEach(() =>
      cy.createService().then(s => {
        svc = s
        return cy.visit(`/services/${s.id}`)
      }),
    )

    it('should display correct information', () => {
      cy.get('body')
        .should('contain', svc.name)
        .should('contain', svc.description)
        .contains('a', svc.ep.name)
        .should('have.attr', 'href', `/escalation-policies/${svc.ep.id}`)
    })

    it('should allow deleting the service', () => {
      cy.pageAction('Delete')
      cy.get('*[role=dialog]')
        .contains('button', 'Confirm')
        .click()
      cy.location('pathname').should('eq', '/services')
      cy.pageSearch(svc.name)
      cy.get('body').should('contain', 'No results')
    })

    it('should handle updating information', () => {
      cy.createEP().then(ep => {
        cy.pageAction('Edit')

        const name = 'SM Svc ' + c.word({ length: 8 })
        const description = c.word({ length: 10 })

        cy.get('input[name=name]')
          .clear()
          .type(name)
        cy.get('textarea[name=description]')
          .clear()
          .type(description)
        cy.get('input[name=escalation-policy]').selectByLabel(ep.name)
        cy.get('*[role=dialog]')
          .find('button[type=submit]')
          .click()

        cy.get('*[role=dialog]').should('not.exist')

        cy.get('body')
          .should('contain', name)
          .should('contain', description)
          .contains('a', ep.name)
          .should('have.attr', 'href', `/escalation-policies/${ep.id}`)
      })
    })

    it('should allow setting and unsetting as a favorite service', () => {
      // test setting as favorite
      cy.get('button[aria-label="Set as a Favorite Service"]').click()
      cy.reload()
      // aria label should change and should be set as a favorite, test unsetting
      cy.get('button[aria-label="Unset as a Favorite Service"').click()
      cy.reload()
      // check that unset
      cy.get('button[aria-label="Set as a Favorite Service"]').click()
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

  describe('Integration Keys', () => {
    let svc: Service
    beforeEach(() =>
      cy.createService().then(s => {
        svc = s
        return cy.visit(`/services/${svc.id}/integration-keys`)
      }),
    )

    const createKey = (type: string, name: string) => {
      cy.pageFab()
      cy.get('input[name=name]').type(name)
      cy.get('input[name=type]').selectByLabel(type)
      cy.get('*[role=dialog]')
        .find('button[type=submit]')
        .click()
      cy.get('*[role=dialog]').should('not.exist')
    }

    it('should allow managing integration keys', () => {
      cy.get('body').should('contain', 'No integration keys')
      ;['Generic API', 'Grafana'].forEach(type => {
        const name = 'SM Int ' + c.word({ length: 8 })
        createKey(type, name)

        cy.get('ul[data-cy=int-keys]').should('contain', name)

        // delete
        cy.get('ul[data-cy=int-keys')
          .contains('li', name)
          .find('button')
          .click()
        cy.get('*[role=dialog]')
          .contains('button', 'Confirm')
          .click()
      })
    })

    it('should manage integration keys with mailgun disabled', () => {
      cy.get('body').should('contain', 'No integration keys')

      const domain = c.domain()
      cy.updateConfig({
        Mailgun: {
          Enable: true,
          APIKey: 'key-' + c.string({ length: 32, pool: '0123456789abcdef' }),
          EmailDomain: domain,
        },
      })
      cy.reload()

      const name = 'SM Int ' + c.word({ length: 8 })
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
      cy.get('input[name=type]')
        .findByLabel('Email')
        .should('not.exist')
    })
  })

  describe('Alerts', () => {
    let svc: Service
    beforeEach(() =>
      cy.createService().then(s => {
        svc = s
        return cy.visit(`/services/${s.id}/alerts`)
      }),
    )

    it('should allow creating alerts', () => {
      cy.pageFab()
      const summary = c.sentence({ words: 3 })
      const details = c.word({ length: 10 })
      cy.get('input[name=summary]').type(summary)
      cy.get('textarea[name=details]').type(details)
      cy.get('input[name=service]').selectByLabel(svc.name)

      cy.get('*[role=dialog]')
        .contains('button', 'Submit')
        .click()

      cy.location('pathname').should('contain', '/alerts/') // details page

      cy.get('body').should('contain', summary)
    })

    it('should allow ack/close all alerts', () => {
      cy.createAlert({ serviceID: svc.id })
      cy.createAlert({ serviceID: svc.id })
      cy.createAlert({ serviceID: svc.id })

      cy.reload()

      cy.get('ul[data-cy=alerts-list]').should('contain', 'UNACKNOWLEDGED')

      cy.pageAction('Acknowledge All')
      cy.get('*[role=dialog]')
        .contains('button', 'Confirm')
        .click()

      cy.get('ul[data-cy=alerts-list]').should('contain', 'ACKNOWLEDGED')
      cy.get('ul[data-cy=alerts-list]').should('not.contain', 'UNACKNOWLEDGED')

      cy.pageAction('Close All')
      cy.get('*[role=dialog]')
        .contains('button', 'Confirm')
        .click()

      cy.get('body').should('contain', 'No results')
    })
  })

  describe('Labels', () => {
    let label: Label
    beforeEach(() =>
      cy.createLabel().then(l => {
        label = l
        return cy.visit(`/services/${l.svcID}/labels`)
      }),
    )

    it('should create a label', () => {
      const key = label.key
      const value = c.word({ length: 10 })

      cy.pageFab()
      cy.get('input[name="key"]').selectByLabel(key)
      cy.get('input[name=value]').type(value)
      cy.get('*[role=dialog]')
        .find('button[type=submit]')
        .click()
      cy.get('li').should('contain', key)
    })

    it('should set an existing label', () => {
      cy.createService().then(svc => {
        cy.visit(`/services/${svc.id}/labels`)
      })

      const newVal = c.word({ length: 10 })
      cy.pageFab()

      cy.get('input[name=key]').selectByLabel(label.key)
      cy.get('input[name=value]').type(newVal)

      cy.get('*[role=dialog]')
        .find('button[type=submit]')
        .click()

      cy.get('li').should('contain', label.key)
      cy.get('li').should('contain', newVal)
    })

    it('should edit a label', () => {
      const newVal = c.word({ length: 10 })

      cy.get('li')
        .should('contain', label.key)
        .find('div')
        .find('button[data-cy=other-actions]')
        .menu('Edit')

      cy.get('input[name=value]')
        .clear()
        .type(newVal)
      cy.get('*[role=dialog]')
        .find('button[type=submit]')
        .click()

      cy.get('li').should('contain', label.key)
      cy.get('li').should('contain', newVal)
    })

    it('should delete a label', () => {
      cy.get('li')
        .should('contain', label.key)
        .find('div')
        .find('button[data-cy=other-actions]')
        .menu('Delete')

      cy.get('*[role=dialog]')
        .find('button[type=submit]')
        .click()

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
      const randomWord = c.word({ min: 5, max: 10 })
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
