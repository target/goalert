import React, { useContext } from 'react'
import { gql, useQuery } from '@apollo/client'
import { ConfigType, ConfigValue, ConfigID } from '../../schema'

type Value = boolean | number | string | string[] | null
export type ConfigData = Record<ConfigID, Value>

const ConfigContext = React.createContext({
  config: [] as ConfigValue[],
  isAdmin: false as boolean,
  userID: null as string | null,
})
ConfigContext.displayName = 'ConfigContext'

const query = gql`
  query {
    user {
      id
      role
    }
    config {
      id
      type
      value
    }
  }
`

type ConfigProviderProps = {
  children: JSX.Element | JSX.Element[]
}

export function ConfigProvider(props: ConfigProviderProps): JSX.Element {
  const { data } = useQuery(query)

  return (
    <ConfigContext.Provider
      value={{
        config: data?.config || [],
        isAdmin: data?.user?.role === 'admin',
        userID: data?.user?.id || null,
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
  userID: string | null
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
    ready: Boolean(ctx.userID),
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

export function Config(props: {
  children: (x: ConfigData, s?: SessionInfo) => JSX.Element
}): JSX.Element {
  return props.children(useConfig(), useSessionInfo()) || null
}

export type RequireConfigProps = {
  configID: ConfigID
  // test to determine whether or not children or else is returned
  test?: (x: Value) => boolean

  // react element to render if checks failed
  else?: JSX.Element
  isAdmin?: boolean
  children: React.ReactChildren
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
