import { GraphQL } from './lib/graphql.ts'
import { login } from './lib/login.ts'

function randomCharacters(length: number): string {
  return Array.from({ length }, () =>
    String.fromCharCode(Math.floor(Math.random() * 26) + 97),
  ).join('')
}

function pickOne<T>(array: T[]): T {
  return array[Math.floor(Math.random() * array.length)]
}

export default function () {
  const token = login()
  const gql = new GraphQL(token)

  for (let i = 0; i < 100; i++) {
    pickOne(gql.rotations).description =
      'k6-random-description-' + randomCharacters(50)
  }
}
