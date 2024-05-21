// import { GraphQLHandler, HttpResponse, graphql } from 'msw'
import {
  ConfigID,
  ConfigType,
  DestinationTypeInfo,
  IntegrationKeyTypeInfo,
  UserRole,
} from '../../schema'
import { destTypes } from './defaultDestTypes'

export type ConfigItem = {
  id: ConfigID
  type: ConfigType
  value: string
}

export type RequireConfigDoc = {
  user: {
    id: string
    name: string
    role: UserRole
  }
  config: ConfigItem[]
  integrationKeyTypes: IntegrationKeyTypeInfo[]
  destinationTypes: DestinationTypeInfo[]
}

export const defaultConfig: RequireConfigDoc = {
  user: {
    id: '00000000-0000-0000-0000-000000000000',
    name: 'Test User',
    role: 'admin',
  },
  config: [
    {
      id: 'General.ApplicationName',
      type: 'string',
      value: 'ComponentTest',
    },
  ],
  integrationKeyTypes: [
    {
      id: 'example-enabled',
      name: 'Enabled Example',
      label: 'Enabled Example Value',
      enabled: true,
    },
    {
      id: 'example-disabled',
      name: 'Disabled Example',
      label: 'Disabled Example Value',
      enabled: false,
    },
  ],
  destinationTypes: destTypes,
}
interface Err {
  message: string
}
type GQLSuccess = { data: object }
type GQLError = { errors: Err[] }

type OpHandler<T> = (vars: T) => GQLSuccess | GQLError

/* mockOp is a helper function that creates a mock for a GraphQL operation that takes an `input` variable (or none) */
export function mockOp<VarType = undefined>(
  operationName: string,
  handler: object | OpHandler<{ input: VarType }>,
): object {
  return {
    matcher: {
      url: 'path:/api/graphql',
      name: `mockOp(${operationName})`,
      body: {
        operationName,
      },
      matchPartialBody: true,
    },
    response: (matcherName: string, req: { body: string }) => {
      const body = JSON.parse(req.body)
      const variables = JSON.parse(req.body).variables || {}
      const result =
        typeof handler === 'function' ? handler(variables) : handler

      console.log(`GraphQL ${operationName}`, {
        body,
        result,
      })

      return result
    },
  }
}

export function mockExpFlags(...flags: string[]): object {
  return mockOp('useExpFlag', {
    data: {
      experimentalFlags: flags,
    },
  })
}

export function mockConfig(config: object): object {
  return mockOp('RequireConfig', {
    data: config,
  })
}
