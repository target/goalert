import { Chance } from 'chance'
import users from '../fixtures/users.json'

const c = new Chance()

declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Creates a new rotation.
       */
      createRotation: typeof createRotation

      /** Delete the rotation with the specified ID */
      deleteRotation: typeof deleteRotation
    }
  }

  type RotationType = 'hourly' | 'daily' | 'weekly'
  interface Rotation {
    id: string
    name: string
    description: string
    timeZone: string
    shiftLength: number
    type: RotationType
    start: string
    users: Array<{
      id: string
      name: string
      email: string
    }>
  }

  interface RotationOptions {
    name?: string
    description?: string
    timeZone?: string
    shiftLength?: number
    type?: RotationType
    start?: string
    favorite?: boolean

    /** Number of participants to add to the rotation. */
    numUsers?: number
  }
}

function createRotation(rot?: RotationOptions): Cypress.Chainable<Rotation> {
  const query = `mutation createRotation($input: CreateRotationInput!){
          createRotation(input: $input) {
            
              id
              name
              description
              timeZone
              shiftLength
              type
              start
              isFavorite
              users {
                id
                name
                email
              }
            
          }
      }`

  if (!rot) rot = {}
  const ids = c.pickset(users, rot.numUsers).map((usr) => usr.id)

  return cy
    .graphql(query, {
      input: {
        name: rot.name || 'SM Rot ' + c.word({ length: 8 }),
        description: rot.description || c.sentence(),
        timeZone: rot.timeZone || 'America/Chicago',
        shiftLength: rot.shiftLength || c.integer({ min: 1, max: 10 }),
        type: rot.type || c.pickone(['hourly', 'daily', 'weekly']),
        start: rot.start
          ? rot.start
          : (c.date({ year: 2017 }) as Date).toISOString(),
        favorite: rot.favorite,
        userIDs: ids,
      },
    })
    .then((res: GraphQLResponse) => {
      const rot = res.createRotation
      return rot
    })
}

function deleteRotation(id: string): Cypress.Chainable<void> {
  const query = `
    mutation deleteAll($input: TargetInput!){
      deleteAll(input: $input) {}
    }
  `

  return cy.graphqlVoid(query, { input: { id, type: 'rotation' } })
}

Cypress.Commands.add('createRotation', createRotation)
Cypress.Commands.add('deleteRotation', deleteRotation)
