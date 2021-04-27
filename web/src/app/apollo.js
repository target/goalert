import {
  ApolloClient,
  InMemoryCache,
  ApolloLink,
  createHttpLink,
} from '@apollo/client'
import { RetryLink } from '@apollo/client/link/retry'
import { authLogout } from './actions'

import reduxStore from './reduxStore'
import { POLL_INTERVAL } from './config'
import promiseBatch from './util/promiseBatch'
import { pathPrefix } from './env'

let pendingMutations = 0
window.onbeforeunload = function (e) {
  if (!pendingMutations) {
    return
  }
  const dialogText =
    'Your changes have not finished saving. If you leave this page, they could be lost.'
  e.returnValue = dialogText
  return dialogText
}

const trackMutation = (p) => {
  pendingMutations++
  p.then(
    () => pendingMutations--,
    () => pendingMutations--,
  )
}

export function doFetch(body, url = pathPrefix + '/api/graphql') {
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

  return promiseBatch(f).then((res) => {
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
    retryIf: (error) => {
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

const graphql2HttpLink = createHttpLink({
  uri: pathPrefix + '/api/graphql',
  fetch: (url, opts) => {
    return doFetch(opts.body, url)
  },
})

const graphql2Link = ApolloLink.from([retryLink, graphql2HttpLink])

const simpleCacheTypes = [
  'alert',
  'rotation',
  'schedule',
  'escalationPolicy',
  'service',
  'user',
  'slackChannel',
  'phoneNumberInfo',
]

// NOTE: see https://www.apollographql.com/docs/react/caching/advanced-topics/#cache-redirects-using-field-policy-read-functions
const typePolicyQueryFields = {}
simpleCacheTypes.forEach((name) => {
  typePolicyQueryFields[name] = {
    read(existingData, { args, toReference, canRead }) {
      return canRead(existingData)
        ? existingData
        : toReference({
            __typename: name,
            id: args?.id,
          })
    },
  }
})

const cache = new InMemoryCache({
  typePolicies: {
    Query: {
      fields: typePolicyQueryFields,
    },
    EscalationPolicy: {
      fields: {
        steps: {
          merge: false,
        },
      },
    },
    Rotation: {
      fields: {
        users: {
          merge: false,
        },
      },
    },
    Schedule: {
      fields: {
        targets: {
          merge: false,
        },
      },
    },
    Service: {
      fields: {
        heartbeatMonitors: {
          merge: false,
        },
        integrationKeys: {
          merge: false,
        },
        labels: {
          merge: false,
        },
      },
    },
    User: {
      fields: {
        calendarSubscriptions: {
          merge: false,
        },
        contactMethods: {
          merge: false,
        },
        notificationRules: {
          merge: false,
        },
        sessions: {
          merge: false,
        },
      },
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
    watchQuery: queryOpts,
  },
})

// errorPolicy can only be set "globally" but breaks if we enable it for existing
// code. Eventually we should transition everything to expect/handle explicit errors.
export const GraphQLClientWithErrors = new ApolloClient({
  link: graphql2Link,
  cache,
  defaultOptions: {
    query: queryOpts,
    watchQuery: queryOpts,
    mutate: { awaitRefetchQueries: true, errorPolicy: 'all' },
  },
})

// refetch all *active* polling queries on mutations
const mutate = GraphQLClient.mutate
GraphQLClient.mutate = (...args) => {
  return mutate.call(GraphQLClient, ...args).then((result) => {
    return Promise.all([
      GraphQLClient.reFetchObservableQueries(true),
      GraphQLClientWithErrors.reFetchObservableQueries(true),
    ]).then(() => result)
  })
}

const mutateWithErrors = GraphQLClientWithErrors.mutate
GraphQLClientWithErrors.mutate = (...args) => {
  return mutateWithErrors
    .call(GraphQLClientWithErrors, ...args)
    .then((result) => {
      return Promise.all([
        GraphQLClient.reFetchObservableQueries(true),
        GraphQLClientWithErrors.reFetchObservableQueries(true),
      ]).then(() => result)
    })
}
