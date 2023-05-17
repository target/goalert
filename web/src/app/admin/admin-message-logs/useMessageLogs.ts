import _ from 'lodash'
import React, { useLayoutEffect, useEffect, useRef, useState } from 'react'
import { gql, useClient } from 'urql'
import { Alert, DebugMessage, MessageLogSearchOptions } from '../../../schema'

export const logsQuery = gql`
  query messageLogsQuery($input: MessageLogSearchOptions) {
    messageLogs(input: $input) {
      nodes {
        id
        createdAt
        updatedAt
        type
        status
        userID
        userName
        source
        destination
        serviceID
        serviceName
        alertID
        providerID
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`

const QUERY_LIMIT = 30

export type MessageLogData = {
  logs: DebugMessage[]
  loading: boolean
  error: Error | undefined
}

// useMessageLogs will fetch up message logs up to the QUERY_LIMIT set above
// using the standard pagination endpoint.
// modeled after useAlerts.ts
export function useMessageLogs(
  options: MessageLogSearchOptions,
  depKey: string,
  pause?: boolean,
): MessageLogData {
  const { createdBefore, createdAfter, omit, search } = options
  const [logs, setLogs] = useState<DebugMessage[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | undefined>()
  const key = useRef(depKey)
  key.current = depKey

  useEffect(() => {
    return () => {
      // cancel on unmount
      key.current = ''
    }
  }, [])

  const client = useClient()
  const fetch = React.useCallback(async () => {
    setLogs([])
    setLoading(true)
    setError(undefined)
    if (pause) {
      return
    }
    async function fetchLogs(
      cursor: string,
    ): Promise<[Alert[], boolean, string, Error | undefined]> {
      const q = await client
        .query(logsQuery, {
          input: {
            first: QUERY_LIMIT,
            after: cursor,
            createdBefore,
            createdAfter,
            omit,
            search,
          },
        })
        .toPromise()

      if (q.error) {
        return [[], false, '', q.error]
      }

      return [
        q.data.messageLogs.nodes,
        q.data.messageLogs.pageInfo.hasNextPage,
        q.data.messageLogs.pageInfo.endCursor,
        undefined,
      ]
    }

    const throttledSetLogs = _.throttle(
      (logs, loading) => {
        setLogs(logs) // todo: additional sorting needed?
        setLoading(loading)
      },
      3000,
      { leading: true },
    )

    let [logs, hasNextPage, endCursor, error] = await fetchLogs('')

    if (key.current !== depKey) return // abort if the key has changed
    if (error) {
      setError(error)
      throttledSetLogs.cancel()
      return
    }
    let allLogs = logs
    throttledSetLogs(allLogs, true)
    while (hasNextPage) {
      ;[logs, hasNextPage, endCursor, error] = await fetchLogs(endCursor)
      if (key.current !== depKey) return // abort if the key has changed
      if (error) {
        setError(error)
        throttledSetLogs.cancel()
        return
      }
      allLogs = allLogs.concat(logs)
      throttledSetLogs(allLogs, true)
    }

    throttledSetLogs(allLogs, false)
  }, [depKey, pause, search])

  useLayoutEffect(() => {
    fetch()
  }, [depKey, pause, search])

  return {
    logs,
    loading,
    error,
  }
}
