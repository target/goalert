import { retryExchange } from '@urql/exchange-retry'
import {
  cacheExchange,
  createClient,
  dedupExchange,
  fetchExchange,
  Exchange,
  Operation,
} from 'urql'
import { pipe, tap } from 'wonka'
import { GraphQLClient, GraphQLClientWithErrors } from './apollo'
import { pathPrefix, isCypress } from './env'

const refetch: Array<(force: boolean) => void> = []
export function refetchAll(force = false): void {
  refetch.forEach((refetch) => refetch(force))
}

Object.defineProperty(window, 'refetchAll', {
  get() {
    return refetchAll
  },
})

// allow refetching all active queries at any time
const refetchExchange = (): Exchange => {
  return ({ client, forward }) =>
    (ops) => {
      const watchedOps$ = new Map<number, Operation>()
      const observedOps$ = new Map<number, number>()

      refetch.push((force) => {
        if (!force) {
          // use existing policy (cache-first will not re-fetch)
          watchedOps$.forEach((op) => {
            client.reexecuteOperation(
              client.createRequestOperation('query', op),
            )
          })
          return
        }

        watchedOps$.forEach((op) => {
          client.reexecuteOperation(
            client.createRequestOperation('query', op, {
              ...op.context,
              requestPolicy: 'cache-and-network',
            }),
          )
        })
      })

      const handleOp = (op: Operation): void => {
        if (op.kind === 'query' && !observedOps$.has(op.key)) {
          observedOps$.set(op.key, 1)
          watchedOps$.set(op.key, op)
        }

        if (op.kind === 'teardown' && observedOps$.has(op.key)) {
          observedOps$.delete(op.key)
          watchedOps$.delete(op.key)
        }
      }

      return forward(pipe(ops, tap(handleOp)))
    }
}

// refetch every 15 sec or on refocus, every 10 sec for Cypress
let poll: NodeJS.Timeout
function resetPoll(): void {
  if (new URLSearchParams(location.search).get('poll') === '0' || isCypress)
    return
  clearInterval(poll)
  poll = setInterval(refetchAll, 15000)
}
window.addEventListener('visibilitychange', () => {
  switch (document.visibilityState) {
    case 'visible':
      resetPoll()
      refetchAll(true)
      break
    case 'hidden':
      clearInterval(poll)
      break
  }
})
resetPoll()

// handle re-fetching Apollo queries on urql mutation
//
// TODO: remove this once apollo is no longer used
const apolloRefetchExchange: Exchange = ({ forward }) => {
  return (operations$) => {
    const operationResult$ = forward(operations$)
    return pipe(
      operationResult$,
      tap((result) => {
        if (result.error) return
        if (result.operation.kind !== 'mutation') return
        GraphQLClient.reFetchObservableQueries(true)
        GraphQLClientWithErrors.reFetchObservableQueries(true)
      }),
    )
  }
}

export const client = createClient({
  url: pathPrefix + '/api/graphql',
  suspense: true,
  exchanges: [
    dedupExchange,
    refetchExchange(),
    cacheExchange,
    apolloRefetchExchange,
    retryExchange({}) as Exchange,
    fetchExchange,
  ],
  requestPolicy: 'cache-and-network',
})
