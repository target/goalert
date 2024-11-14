import { Entity } from './entity.ts'
import { GraphQL } from './graphql.ts'
import {
  randDate,
  randInt,
  randPickOne,
  randSample,
  randString,
  randTimeZone,
} from './rand.ts'

type RotationType = 'daily' | 'weekly' | 'hourly'

export type RotationParams = {
  name: string
  description: string
  timeZone: string
  type: RotationType
  shiftLength: number
  start: string
  userIDs: string[]
}

export class Rotation extends Entity {
  constructor(gql: GraphQL, id: string) {
    super(gql, 'rotation', id)
  }

  static randParams(gql: GraphQL): RotationParams {
    return {
      name: randString(50),
      description: randString(128),
      timeZone: randTimeZone(),
      type: randPickOne(['daily', 'weekly', 'hourly']),
      shiftLength: randInt(1, 24),
      start: randDate().toISOString(),
      userIDs: randSample(gql.users, 20).map((u) => u.id),
    }
  }

  public get timeZone(): string {
    return this.getField('timeZone') as string
  }

  public set timeZone(value: string) {
    this.setField('timeZone', value)
  }

  public get type(): RotationType {
    return this.getField('type') as RotationType
  }

  public set type(value: RotationType) {
    this.setField('type', value, 'RotationType')
  }

  public get shiftLength(): number {
    return this.getField('shiftLength') as number
  }

  public set shiftLength(value: number) {
    this.setField('shiftLength', value)
  }

  public get start(): string {
    return this.getField('start') as string
  }

  public set start(value: string) {
    this.setField('start', value)
  }
}
