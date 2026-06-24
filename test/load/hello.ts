import { login } from './lib/login.ts'
import { GraphQL } from './lib/graphql.ts'
import { check } from 'k6'

export default function (): void {
  const gql = new GraphQL(login())

  check(gql.userID, {
    'userID is not empty': (u) => u !== '',
  })
}
