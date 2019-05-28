import { Chance } from 'chance'

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

    /** Number of participants to add to the rotation. */
    count?: number
  }
}

type Part = { pos: number }
const sortParts = (a: Part, b: Part) => (a.pos < b.pos ? -1 : 1)

function createRotation(rot?: RotationOptions): Cypress.Chainable<Rotation> {
  const query = `mutation createRotation($input: CreateAllInput!){
          createAll(input: $input) {
            rotations {
              id
              name
              description
              timeZone: time_zone
              shiftLength: shift_length
              type
              start
              users: participants {
                pos: position
                user {
                  id
                  name
                  email
                }
              }
            }
          }
      }`

  return cy.fixture('users').then(users => {
    if (!rot) rot = {}
    const ids = c.pickset(users, rot.count).map((usr: any) => usr.id)
    const parts = ids.map(id => ({ rotation_id: 'rot', user_id: id }))

    return cy
      .graphql(query, {
        input: {
          rotations: [
            {
              id_placeholder: 'rot',
              name: rot.name || 'SM Rot ' + c.word({ length: 8 }),
              description: rot.description || c.sentence(),
              time_zone: rot.timeZone || 'America/Chicago',
              shift_length: rot.shiftLength || c.integer({ min: 1, max: 10 }),
              type: rot.type || c.pickone(['hourly', 'daily', 'weekly']),
              start: rot.start
                ? rot.start
                : (c.date({ year: 2017 }) as Date).toISOString(),
            },
          ],
          rotation_participants: parts,
        },
      })
      .then(res => {
        const rot = res.createAll.rotations[0]
        rot.users.sort(sortParts)
        rot.users = rot.users.map((u: any) => u.user)
        return rot
      })
  })
}

function deleteRotation(id: string): Cypress.Chainable<void> {
  const query = `
    mutation deleteRotation($input: DeleteRotationInput!){
      deleteRotation(input: $input) { deleted_id }
    }
  `

  return cy.graphql(query, { input: { id } })
}

Cypress.Commands.add('createRotation', createRotation)
Cypress.Commands.add('deleteRotation', deleteRotation)
