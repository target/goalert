import { Chance } from 'chance'

const c = new Chance()

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

  return cy.fixture('users').then((users: Profile[]) => {
    if (!rot) rot = {}
    const ids = c.pickset(users, rot.count).map((usr: Profile) => usr.id)

    return cy
      .graphql2(query, {
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
      .then(res => {
        const rot = res.createRotation
        return rot
      })
  })
}

function deleteRotation(id: string): Cypress.Chainable<void> {
  const query = `
    mutation deleteAll($input: TargetInput!){
      deleteAll(input: $input) {}
    }
  `

  return cy.graphql2(query, { input: { id: id, type: 'rotation' } })
}

Cypress.Commands.add('createRotation', createRotation)
Cypress.Commands.add('deleteRotation', deleteRotation)
