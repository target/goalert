import { GraphQL } from './graphql.ts'

type EntityBase = {
  id: string
  name: string
  description: string
  isFavorite: boolean
}

function updateMutName(typeName: string): string {
  return 'update' + typeName.charAt(0).toUpperCase() + typeName.slice(1)
}

// Entity is a base class for all entities in the system (e.g., schedules, rotations, etc.)
export class Entity {
  constructor(
    private gql: GraphQL,
    private typeName: string,
    public readonly id: string,
  ) {}

  private get base(): EntityBase {
    return this.gql.query(
      `query GetBase($id: ID!){base: ${this.typeName}(id: $id){id name description isFavorite}}`,
      { id: this.id },
    ).base
  }

  // This method is used to set a string field on the entity.
  //
  // Example: to set the name of a rotation, you would call `setStringField('name', 'new name')`.
  private setStringField(fieldName: string, value: string): void {
    this.gql.query(
      `mutation SetField($id: ID!, $value: String!){${updateMutName(this.typeName)}(input: {id: $id, ${fieldName}: $value})}`,
      { id: this.id, value },
    )
  }

  public get name(): string {
    return this.base.name
  }

  public set name(value: string) {
    this.setStringField('name', value)
  }

  public get description(): string {
    return this.base.description
  }

  public set description(value: string) {
    this.setStringField('description', value)
  }

  public get isFavorite(): boolean {
    return this.base.isFavorite
  }

  public set isFavorite(value: boolean) {
    this.gql.query(
      `mutation SetIsFavorite($tgt: TargetInput!, $value: Boolean!){setFavorite(input: {target: $tgt, isFavorite: $value})}`,
      { tgt: { id: this.id, type: this.typeName }, value },
    )
  }
}
