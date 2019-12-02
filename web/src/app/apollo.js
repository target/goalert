import ApolloClient from 'apollo-client'
import { ApolloLink } from 'apollo-link'
import { createHttpLink } from 'apollo-link-http'
import { RetryLink } from 'apollo-link-retry'
import { InMemoryCache } from 'apollo-cache-inmemory'
import { camelCase } from 'lodash-es'
import { toIdValue } from 'apollo-utilities'
import { authLogout } from './actions'

import reduxStore from './reduxStore'
import { POLL_INTERVAL } from './config'

let pendingMutations = 0
window.onbeforeunload = function(e) {
  if (!pendingMutations) {
    return
  }
  const dialogText =
    'Your changes have not finished saving. If you leave this page, they could be lost.'
  e.returnValue = dialogText
  return dialogText
}

const trackMutation = p => {
  pendingMutations++
  p.then(
    () => pendingMutations--,
    () => pendingMutations--,
  )
}

export function doFetch(body, url = '/v1/graphql') {
  const f = fetch(url, {
    credentials: 'same-origin',
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body,
  })

  if (body.query && body.query.startsWith && body.query.startsWith('mutation'))
    trackMutation(f)

  return f.then(res => {
    if (res.ok) {
      return res
    }

    if (res.status === 401) {
      reduxStore.dispatch(authLogout())
    }

    throw new Error('HTTP Response ' + res.status + ': ' + res.statusText)
  })
}

const retryLink = new RetryLink({
  delay: {
    initial: 500,
    max: 3000,
    jitter: true,
  },
  attempts: {
    max: 5,
    retryIf: (error, _operation) => {
      // Retry on any error except HTTP Response errors with the
      // exception of 502-504 response codes (e.g. no retry on 401/auth etc..).
      return (
        !!error &&
        (!/^HTTP Response \d+:/.test(error.message) ||
          /^HTTP Response 50[234]:/.test(error.message))
      )
    },
  },
})

const defaultHttpLink = createHttpLink({
  uri: '/v1/graphql',
  fetch: (url, opts) => {
    return doFetch(opts.body)
  },
})

// compose links
const defaultLink = ApolloLink.from([
  retryLink,
  defaultHttpLink, // terminating link must be last: apollographql.com/docs/link/overview.html#terminating
])

export const LegacyGraphQLClient = new ApolloClient({
  link: defaultLink,
  cache: new InMemoryCache(),
  defaultOptions: { errorPolicy: 'all' },
})

const graphql2HttpLink = createHttpLink({
  uri: '/api/graphql',
  fetch: (url, opts) => {
    return doFetch(opts.body, url)
  },
})

const graphql2Link = ApolloLink.from([retryLink, graphql2HttpLink])

const simpleCacheTypes = [
  'Alert',
  'Rotation',
  'Schedule',
  'EscalationPolicy',
  'Service',
  'User',
  'SlackChannel',
]

// tell Apollo to use cached data for `type(id: foo) {... }` queries
const queryCache = {}

simpleCacheTypes.forEach(name => {
  queryCache[camelCase(name)] = (_, args) =>
    args &&
    toIdValue(
      cache.config.dataIdFromObject({
        __typename: name,
        id: args.id,
      }),
    )
})

const cache = new InMemoryCache({
  cacheRedirects: {
    Query: {
      ...queryCache,
    },
  },
})

const queryOpts = { fetchPolicy: 'cache-and-network', errorPolicy: 'all' }
if (new URLSearchParams(location.search).get('poll') !== '0') {
  queryOpts.pollInterval = POLL_INTERVAL
}

export const GraphQLClient = new ApolloClient({
  link: graphql2Link,
  cache,
  defaultOptions: {
    query: queryOpts,
    mutate: { awaitRefetchQueries: true },
  },
})

// errorPolicy can only be set "globally" but breaks if we enable it for existing
// code. Eventually we should transition everything to expect/handle explicit errors.
export const GraphQLClientWithErrors = new ApolloClient({
  link: graphql2Link,
  cache,
  defaultOptions: {
    query: queryOpts,
    mutate: { awaitRefetchQueries: true, errorPolicy: 'all' },
  },
})
