import { GraphQL } from './graphql.ts'

function updateMutName(typeName: string): string {
  return 'update' + typeName.charAt(0).toUpperCase() + typeName.slice(1)
}

function guessType(val: unknown): string {
  switch (typeof val) {
    case 'string':
      return 'String'
    case 'number':
      return 'Int'
    case 'boolean':
      return 'Boolean'
    default:
      throw new Error(`Unknown type: ${typeof val}`)
  }
}

// Entity is a base class for all entities in the system (e.g., schedules, rotations, users, etc.)
export class BaseEntity {
  constructor(
    private gql: GraphQL,
    private typeName: string,
    public readonly id: string,
  ) {}

  protected getField(fieldName: string): unknown {
    return this.gql.query(
      `query GetField($id: ID!){base: ${this.typeName}(id: $id){value: ${fieldName}}}`,
      { id: this.id },
    ).base.value
  }

  // This method is used to set a field on the entity.
  //
  // Example: to set the name of a rotation, you would call `setField('name', 'new name')`.
  protected setField(
    fieldName: string,
    value: unknown,
    typeName: string = guessType(value),
  ): void {
    this.gql.query(
      `mutation SetField($id: ID!, $value: ${typeName}!){${updateMutName(this.typeName)}(input: {id: $id, ${fieldName}: $value})}`,
      { id: this.id, value },
    )
  }

  public delete(): void {
    this.gql.query(
      `mutation Delete($tgt: TargetInput!){deleteAll(input: [$tgt])}`,
      {
        tgt: {
          type: this.typeName,
          id: this.id,
        },
      },
    )
  }

  public get name(): string {
    return this.getField('name') as string
  }

  public set name(value: string) {
    this.setField('name', value)
  }

  public get isFavorite(): boolean {
    return this.getField('isFavorite') as boolean
  }

  public set isFavorite(value: boolean) {
    this.gql.query(
      `mutation SetIsFavorite($tgt: TargetInput!, $value: Boolean!){setFavorite(input: {target: $tgt, isFavorite: $value})}`,
      { tgt: { id: this.id, type: this.typeName }, value },
    )
  }
}

export class Entity extends BaseEntity {
  constructor(gql: GraphQL, typeName: string, id: string) {
    super(gql, typeName, id)
  }

  public get description(): string {
    return this.getField('description') as string
  }

  public set description(value: string) {
    this.setField('description', value)
  }
}
