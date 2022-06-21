import _ from 'lodash'
import React, { useLayoutEffect } from 'react'
import { useEffect, useRef, useState } from 'react'
import { gql, useClient } from 'urql'
import { Alert } from '../../../schema'

const alertsQuery = gql`
  query alerts($input: AlertSearchOptions!) {
    alerts(input: $input) {
      nodes {
        id
        alertID
        summary
        status
        service {
          name
          id
        }
        createdAt
        metrics {
          closedAt
          timeToClose
          timeToAck
          escalated
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`

const QUERY_LIMIT = 1000

export type AlertsData = {
  alerts: Alert[]
  loading: boolean
  error: Error | undefined
}

export function useAlerts(
  serviceID: string,
  since: string,
  until: string,
  pause?: boolean,
): AlertsData {
  const depKey = `${serviceID}-${since}-${until}`
  const [alerts, setAlerts] = useState<Alert[]>([])
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
    setAlerts([])
    setLoading(true)
    setError(undefined)
    if (pause) {
      return
    }
    async function fetchAlerts(
      cursor: string,
    ): Promise<[Alert[], boolean, string, Error | undefined]> {
      const q = await client
        .query(alertsQuery, {
          input: {
            filterByServiceID: [serviceID],
            first: QUERY_LIMIT,
            notClosedBefore: since,
            closedBefore: until,
            filterByStatus: ['StatusClosed'],
            after: cursor,
          },
        })
        .toPromise()

      if (q.error) {
        return [[], false, '', q.error]
      }

      return [
        q.data.alerts.nodes,
        q.data.alerts.pageInfo.hasNextPage,
        q.data.alerts.pageInfo.endCursor,
        undefined,
      ]
    }

    const throttledSetAlerts = _.throttle(
      (alerts, loading) => {
        setAlerts(_.sortBy(alerts, 'metrics.closedAt'))
        setLoading(loading)
      },
      3000,
      { leading: true },
    )

    let [alerts, hasNextPage, endCursor, error] = await fetchAlerts('')
    if (key.current !== depKey) return // abort if the key has changed
    if (error) {
      setError(error)
      throttledSetAlerts.cancel()
      return
    }
    let allAlerts = alerts
    throttledSetAlerts(allAlerts, true)
    while (hasNextPage) {
      ;[alerts, hasNextPage, endCursor, error] = await fetchAlerts(endCursor)
      if (key.current !== depKey) return // abort if the key has changed
      if (error) {
        setError(error)
        throttledSetAlerts.cancel()
        return
      }
      allAlerts = allAlerts.concat(alerts)
      throttledSetAlerts(allAlerts, true)
    }

    throttledSetAlerts(allAlerts, false)
  }, [depKey, pause])

  useLayoutEffect(() => {
    fetch()
  }, [depKey, pause])

  return {
    alerts,
    loading,
    error,
  }
}
