import {
  cacheExchange,
  createClient,
  dedupExchange,
  fetchExchange,
  Exchange,
  Operation,
} from 'urql'
import { pipe, tap } from 'wonka'
import { pathPrefix } from './env'

const refetch: Array<() => void> = []
export function refetchAll() {
  refetch.forEach((refetch) => refetch())
}

// allow refetching all active queries at any time
const refetchExchange = (): Exchange => {
  return ({ client, forward }) =>
    (ops) => {
      const watchedOps$ = new Map<number, Operation>()
      const observedOps$ = new Map<number, number>()

      refetch.push(() => {
        watchedOps$.forEach((op) => {
          client.reexecuteOperation(
            client.createRequestOperation('query', op, {
              ...op.context,
              requestPolicy: 'cache-and-network',
            }),
          )
        })
      })

      const handleOp = (op: Operation) => {
        if (
          op.kind === 'query' &&
          !observedOps$.has(op.key) &&
          !op.context.noPoll
        ) {
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

// refetch every 15 sec or on refocus
let poll: NodeJS.Timer
function resetPoll() {
  if (new URLSearchParams(location.search).get('poll') === '0') return
  clearInterval(poll)
  poll = setInterval(refetchAll, 15000)
}
window.addEventListener('visibilitychange', () => {
  switch (document.visibilityState) {
    case 'visible':
      resetPoll()
      refetchAll()
      break
    case 'hidden':
      clearInterval(poll)
      break
  }
})
resetPoll()

export const client = createClient({
  url: pathPrefix + '/api/graphql',
  exchanges: [dedupExchange, refetchExchange(), cacheExchange, fetchExchange],
  requestPolicy: 'cache-and-network',
})
