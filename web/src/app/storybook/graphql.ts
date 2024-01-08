import { GraphQLHandler, HttpResponse, graphql } from 'msw'
import {
  ConfigID,
  ConfigType,
  IntegrationKeyTypeInfo,
  UserRole,
} from '../../schema'

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
}

export function handleConfig(doc: RequireConfigDoc): GraphQLHandler {
  return graphql.query('RequireConfig', () => {
    return HttpResponse.json({
      data: doc,
    })
  })
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
}

export const handleDefaultConfig = handleConfig(defaultConfig)
