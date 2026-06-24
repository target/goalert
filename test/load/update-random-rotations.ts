import { GraphQL } from './lib/graphql.ts'
import { login } from './lib/login.ts'
import { randString } from './lib/rand.ts'

export default function (): void {
  const adminGQL = new GraphQL(login())
  const [gql, testUser] = adminGQL.newAdmin()

  for (let i = 0; i < 10; i++) {
    const rot = gql.genRotation()
    rot.type = 'daily' // update type
    for (let j = 0; j < 10; j++) {
      rot.description = 'update-desc-' + randString(128) // update description
      rot.name = 'update-name-' + randString(40)
      rot.shiftLength++ // update shiftLength
      rot.description = randString(128) // update description
    }
    rot.delete()
  }

  testUser.delete()
}
