import { Chance } from 'chance'
import { testScreen } from '../support'
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
      names = []
      nameSubstr = c.word({ length: 12 })
      for (let i = 0; i < 45; i++) {
        const name = `Pagination Test ${nameSubstr} ${padZeros('' + i)}`
        names.push(name)
      }

      return create(names.map((name) => ({ name })))
    })

    beforeEach(() => cy.visit(`/${url}?search=Test+${nameSubstr}`))

    it(`should load more list items when scrolling to the bottom`, () => {
      for (let i = 0; i < itemsPerPage; i++)
        cy.get('body').should('contain', names[i])
      cy.get('[id="content"]').scrollTo('bottom')
        cy.get('[data-cy=apollo-list] li a').should('have.length', 30)
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
