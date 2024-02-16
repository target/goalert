import React, { ReactChild, useContext } from 'react'
import { gql, useQuery } from 'urql'
import {
  ConfigType,
  ConfigValue,
  ConfigID,
  IntegrationKeyTypeInfo,
  DestinationTypeInfo,
  DestinationType,
} from '../../schema'
import { useExpFlag } from './useExpFlag'

type Value = boolean | number | string | string[] | null
export type ConfigData = Record<ConfigID, Value>

const ConfigContext = React.createContext({
  integrationKeyTypes: [] as IntegrationKeyTypeInfo[],
  config: [] as ConfigValue[],
  isAdmin: false as boolean,
  userID: '' as string,
  userName: null as string | null,
  destTypes: [] as DestinationTypeInfo[],
})
ConfigContext.displayName = 'ConfigContext'

const query = gql`
  query RequireConfig {
    user {
      id
      name
      role
    }
    config {
      id
      type
      value
    }
    integrationKeyTypes {
      id
      name
      label
      enabled
    }
  }
`

// expDestQuery will be used when "dest-types" experimental flag is enabled.
// expDestQuery should replace "query" when new components have been fully integrated into master branch for https://github.com/target/goalert/issues/2639
const expDestQuery = gql`
  query RequireConfig {
    user {
      id
      name
      role
    }
    config {
      id
      type
      value
    }
    integrationKeyTypes {
      id
      name
      label
      enabled
    }
    destinationTypes {
      type
      name
      enabled
      disabledMessage
      userDisclaimer

      isContactMethod
      isEPTarget
      isSchedOnCallNotify

      requiredFields {
        fieldID
        labelSingular
        labelPlural
        hint
        hintURL
        placeholderText
        prefix
        inputType
        isSearchSelectable
        supportsValidation
      }
    }
  }
`

type ConfigProviderProps = {
  children: ReactChild | ReactChild[]
}

export function ConfigProvider(props: ConfigProviderProps): React.ReactNode {
  const hasDestTypesFlag = useExpFlag('dest-types')
  const [{ data }] = useQuery({
    query: hasDestTypesFlag ? expDestQuery : query,
  })

  return (
    <ConfigContext.Provider
      value={{
        integrationKeyTypes: data?.integrationKeyTypes || [],
        config: data?.config || [],
        isAdmin: data?.user?.role === 'admin',
        userID: data?.user?.id || null,
        userName: data?.user?.name || null,
        destTypes: data?.destinationTypes || [],
      }}
    >
      {props.children}
    </ConfigContext.Provider>
  )
}

function parseValue(type: ConfigType, value: string): Value {
  if (!type) return null
  switch (type) {
    case 'boolean':
      return value === 'true'
    case 'integer':
      return parseInt(value, 10)
    case 'string':
      return value
    case 'stringList':
      return value === '' ? [] : value.split('\n')
  }

  throw new TypeError(`unknown config type '${type}'`)
}

function isTrue(value: Value): boolean {
  if (Array.isArray(value)) return value.length > 0
  if (value === 'false') return false
  return Boolean(value)
}

const mapConfig = (value: ConfigValue[]): ConfigData => {
  const data: { [x: string]: Value } = {}
  value.forEach((v) => {
    data[v.id] = parseValue(v.type, v.value)
  })
  return data as ConfigData
}

export type SessionInfo = {
  isAdmin: boolean
  userID: string
  userName: string | null
  ready: boolean
}

// useSessionInfo returns an object with the following properties:
// - `isAdmin` true if the current session is an admin
// - `userID` the current users ID
// - `ready` true if session/config info is available (e.g. before initial page load/fetch)
export function useSessionInfo(): SessionInfo {
  const ctx = useContext(ConfigContext)

  return {
    isAdmin: ctx.isAdmin,
    userID: ctx.userID,
    userName: ctx.userName,
    ready: Boolean(ctx.userID),
  }
}

export type FeatureInfo = {
  // integrationKeyTypes is a list of integration key types available in the
  // current environment.
  integrationKeyTypes: IntegrationKeyTypeInfo[]
}

// useFeatures returns information about features available in the current
// environment.
export function useFeatures(): FeatureInfo {
  const ctx = useContext(ConfigContext)

  return {
    integrationKeyTypes: ctx.integrationKeyTypes,
  }
}

// useConfig will return the current public configuration as an object
// like:
//
// ```js
// {
//   "Mailgun.Enable": true
// }
// ```
export function useConfig(): ConfigData {
  return mapConfig(useContext(ConfigContext).config)
}

// useConfigValue will return an array of config values
// for the provided fields.
//
// Example:
// ```js
// const [mailgun, slack] = useConfigValue('Mailgun.Enable', 'Slack.Enable')
// ```
export function useConfigValue(...fields: ConfigID[]): Value[] {
  const config = useConfig()
  return fields.map((f) => config[f])
}

// useContactMethodTypes returns a list of contact method destination types.
export function useContactMethodTypes(): DestinationTypeInfo[] {
  const cfg = React.useContext(ConfigContext)
  return cfg.destTypes.filter((t) => t.isContactMethod)
}

// useDestinationType returns information about the given destination type.
export function useDestinationType(type: DestinationType): DestinationTypeInfo {
  const ctx = React.useContext(ConfigContext)
  const typeInfo = ctx.destTypes.find((t) => t.type === type)

  if (!typeInfo) throw new Error(`unknown destination type '${type}'`)

  return typeInfo
}

export function Config(props: {
  children: (x: ConfigData, s?: SessionInfo) => JSX.Element
}): JSX.Element {
  return props.children(useConfig(), useSessionInfo()) || null
}

export type RequireConfigProps = {
  configID?: ConfigID
  // test to determine whether or not children or else is returned
  test?: (x: Value) => boolean

  // react element to render if checks failed
  else?: JSX.Element
  isAdmin?: boolean
  children?: ReactChild
}

export default function RequireConfig(
  props: RequireConfigProps,
): JSX.Element | null {
  const {
    configID,
    test = isTrue,
    isAdmin: wantIsAdmin,
    children,
    else: elseValue = null,
  } = props
  const config = useConfig()
  const { isAdmin } = useSessionInfo()

  if (wantIsAdmin && !isAdmin) {
    return elseValue
  }

  if (configID && !test(config[configID])) {
    return elseValue
  }

  return <React.Fragment>{children}</React.Fragment>
}
