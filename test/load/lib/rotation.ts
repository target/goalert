import { Entity } from './entity.ts'
import { GraphQL } from './graphql.ts'

export class Rotation extends Entity {
  constructor(gql: GraphQL, id: string) {
    super(gql, 'rotation', id)
  }
}
