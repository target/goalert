import { Chance } from 'chance'
import { testScreen, login } from '../support/e2e'
const c = new Chance()

const itemsPerPage = 15

const padZeros = (val: string): string => {
  while (val.length < 4) val = '0' + val
  return val
}

interface CreateOpts {
  name: string
}
type createOneFunc = (opts: CreateOpts) => Cypress.Chainable
type createManyFunc = (names: Array<CreateOpts>) => Cypress.Chainable

function testPaginating(
  label: string,
  url: string,
  create: createManyFunc,
): void {
  let names: Array<string> = []
  let nameSubstr = ''

  describe(label, () => {
    before(() => {
      login()

      names = []
      nameSubstr = c.word({ length: 12 })
      for (let i = 0; i < 45; i++) {
        const name = `Pagination Test ${nameSubstr} ${padZeros('' + i)}`
        names.push(name)
      }

      return create(names.map((name) => ({ name })))
    })

    beforeEach(() => cy.visit(`/${url}?search=Test+${nameSubstr}`))

    it(`should navigate forward and back`, () => {
      cy.get('button[data-cy="back-button"]').should('be.disabled')
      for (let i = 0; i < itemsPerPage; i++)
        cy.get('body').should('contain', names[i])

      cy.get('button[data-cy="next-button"]').first().click()

      for (let i = 0; i < itemsPerPage; i++)
        cy.get('body').should('contain', names[itemsPerPage + i])

      cy.get('button[data-cy="next-button"]').last().click()

      for (let i = 0; i < itemsPerPage; i++)
        cy.get('body').should('contain', names[itemsPerPage * 2 + i])

      cy.get('button[data-cy="next-button"]').should('be.disabled')

      cy.get('button[data-cy="back-button"]').first().click()

      for (let i = 0; i < itemsPerPage; i++)
        cy.get('body').should('contain', names[itemsPerPage + i])

      cy.get('button[data-cy="back-button"]').last().click()

      for (let i = 0; i < itemsPerPage; i++)
        cy.get('body').should('contain', names[i])
    })

    it(`should reset to page 0 on search change`, () => {
      cy.get('button[data-cy="back-button"]').should('be.disabled')
      for (let i = 0; i < itemsPerPage; i++)
        cy.get('body').should('contain', names[i])

      cy.get('button[data-cy="next-button"]').first().click()

      for (let i = 0; i < itemsPerPage; i++)
        cy.get('body').should('contain', names[itemsPerPage + i])

      cy.pageSearch(nameSubstr) // trigger search onChange

      cy.get('button[data-cy="back-button"]').should('be.disabled')
      for (let i = 0; i < itemsPerPage; i++)
        cy.get('body').should('contain', names[i])
    })

    it(`should reset search on main nav click`, () => {
      cy.get('button[data-cy="back-button"]').should('be.disabled')
      for (let i = 0; i < itemsPerPage; i++)
        cy.get('body').should('contain', names[i])

      cy.get('button[data-cy="next-button"]').first().click()

      for (let i = 0; i < itemsPerPage; i++)
        cy.get('body').should('contain', names[itemsPerPage + i])
      cy.url().should('contain', 'search')

      cy.pageNav(label)

      cy.get('button[data-cy="back-button"]').should('be.disabled')
      cy.url().should('not.contain', 'search')
    })
  })
}

function createOne(fn: createOneFunc) {
  return (names: Array<CreateOpts>) => {
    names.forEach((name) => fn(name))
    return cy
  }
}

function testPagination(): void {
  testPaginating('Rotations', 'rotations', createOne(cy.createRotation))
  testPaginating('Schedules', 'schedules', createOne(cy.createSchedule))
  testPaginating(
    'Escalation Policies',
    'escalation-policies',
    createOne(cy.createEP),
  )
  testPaginating('Services', 'services', createOne(cy.createService))
  testPaginating('Users', 'users', cy.createManyUsers)
}

testScreen('Pagination', testPagination)
