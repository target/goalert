import { gql, useQuery } from '@apollo/client'
import React, { useContext } from 'react'
import p from 'prop-types'
import _ from 'lodash-es'

const ConfigContext = React.createContext({
  config: [],
  isAdmin: false,
  userID: null,
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

export function ConfigProvider({ children }) {
  const { data } = useQuery(query)

  return (
    <ConfigContext.Provider
      value={{
        config: data?.config || [],
        isAdmin: data?.user?.role === 'admin',
        userID: data?.user?.id || null,
      }}
    >
      {children}
    </ConfigContext.Provider>
  )
}

function parseValue(type, value) {
  if (!type) return null
  switch (type) {
    case 'boolean':
      return value === 'true'
    case 'integer':
      return parseInt(value, 10)
    case 'string':
      return value
    case 'stringList':
      if (value === '') return []
      return value.split('\n')
  }

  throw new TypeError(`unknown config type '${type}'`)
}

function isTrue(value) {
  if (Array.isArray(value)) return value.length > 0

  return Boolean(value)
}

const mapConfig = (value) =>
  _.chain(value)
    .groupBy('id')
    .mapValues((v) => parseValue(v[0].type, v[0].value))
    .value()

// useSessionInfo returns an object with the following properties:
// - `isAdmin` true if the current session is an admin
// - `userID` the current users ID
// - `ready` true if session/config info is available (e.g. before initial page load/fetch)
export function useSessionInfo() {
  const info = _.pick(useContext(ConfigContext), 'isAdmin', 'userID')
  return {
    userID: info.userID,
    isAdmin: info.isAdmin,
    ready: Boolean(info.userID), // no user ID if not loaded
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
export function useConfig() {
  return mapConfig(useContext(ConfigContext).config)
}

// useConfigValue will return an array of config values
// for the provided fields.
//
// Example:
// ```js
// const [mailgun, slack] = useConfigValue('Mailgun.Enable', 'Slack.Enable')
// ```
export function useConfigValue(...fields) {
  const config = useConfig()
  return fields.map((f) => config[f])
}

export function Config({ children }) {
  const { isAdmin, userID, config } = useContext(ConfigContext)
  return children(mapConfig(config), { isAdmin, userID }) || null
}

export default function RequireConfig(props) {
  const {
    configID,
    test = isTrue,
    isAdmin: wantIsAdmin,
    children,
    else: elseValue = null,
    ...rest
  } = props
  const { config, isAdmin } = useContext(ConfigContext)

  if (wantIsAdmin && !isAdmin) {
    return elseValue
  }
  if (configID && !test(config[configID])) {
    return elseValue
  }

  return React.Children.map(children, (child) =>
    React.cloneElement(child, _.omit(rest, Object.keys(child.props))),
  )
}
RequireConfig.propTypes = {
  isAdmin: p.bool,
  configID: p.string,
  test: p.func, // test to determine whether or not else is returned

  else: p.node, // react element to render if checks failed

  children: p.node, // elements to return if checks pass
}
