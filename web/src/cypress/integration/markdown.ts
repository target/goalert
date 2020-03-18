import { Chance } from 'chance'
import { testScreen } from '../support'
const c = new Chance()

testScreen('Markdown', testMarkdown)

interface Config {
  // test is the name of the section
  test: string

  // url is the top-level url path
  url: string

  // gen is a function that generates a thing with a given description.
  gen: (desc: string) => Cypress.Chainable<{ id: string; name: string }>
}

function testMarkdown(screen: ScreenFormat) {
  ;[
    {
      test: 'Rotations',
      url: 'rotations',
      gen: (desc: string) => cy.createRotation({ description: desc }),
    },
    {
      test: 'Schedules',
      url: 'on_call_schedules',
      gen: (desc: string) => cy.createSchedule({ description: desc }),
    },
    {
      test: 'Escalation Policies',
      url: 'escalation_policies',
      gen: (desc: string) => cy.createEP({ description: desc }),
    },
    {
      test: 'Services',
      url: 'services',
      gen: (desc: string) => cy.createService({ description: desc }),
    },
  ].forEach(cfg => describe(cfg.test, () => testMarkdownDesc(screen, cfg)))
}

function testMarkdownDesc(screen: ScreenFormat, cfg: Config) {
  let rawDescription: string
  let generated: { id: string; name: string }
  beforeEach(() => {
    rawDescription = c.word({ length: 8 })
    return cfg.gen('# ' + rawDescription).then(g => {
      generated = g
    })
  })

  describe('Details', () => {
    beforeEach(() => {
      return cy.visit(`/${cfg.url}/${generated.id}`)
    })
    it('should render description with markdown', () => {
      // make sure markdown renders properly
      cy.get('h1')
        .should('not.contain', '#')
        .should('contain', rawDescription)
    })
  })

  describe('List', () => {
    beforeEach(() => {
      cy.visit(`/${cfg.url}`)
      return cy.pageSearch(generated.name)
    })
    it('should render the first line of the description', () => {
      cy.get('body').should('contain', rawDescription)
    })
  })
}
