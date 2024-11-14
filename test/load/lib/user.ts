import { BaseEntity } from './entity.ts'
import { GraphQL } from './graphql.ts'
import { randEmail, randPickOne, randString } from './rand.ts'

type UserRole = 'admin' | 'user'

export type UserParams = {
  name: string
  email: string
  role: UserRole
  username: string
  password: string
}

export class User extends BaseEntity {
  constructor(gql: GraphQL, id: string) {
    super(gql, 'user', id)
  }

  static randParams(gql: GraphQL): UserParams {
    return {
      name: 'user-' + randString(20),
      email: randEmail(),
      role: randPickOne(['admin', 'user']),
      username: 'test-' + randString(16),
      password: randString(20),
    }
  }
}
