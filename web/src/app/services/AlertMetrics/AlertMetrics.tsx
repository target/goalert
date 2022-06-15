import React, {
  useMemo,
  useState,
  useEffect,
  useRef,
  useDeferredValue,
} from 'react'
import { Card, CardContent, CardHeader, Grid } from '@mui/material'
import { gql, useClient } from 'urql'
import { DateTime, Duration, Interval } from 'luxon'
import { useURLParams } from '../../actions/hooks'
import AlertMetricsFilter, {
  DATE_FORMAT,
  MAX_DAY_COUNT,
} from './AlertMetricsFilter'
import AlertCountGraph from './AlertCountGraph'
import AlertMetricsTable from './AlertMetricsTable'
import AlertAveragesGraph from './AlertAveragesGraph'
import { Alert } from '../../../schema'
import { GenericError } from '../../error-pages'
import _ from 'lodash'

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

const QUERY_LIMIT = 100

export type AlertMetricsProps = {
  serviceID: string
}

type AlertsData = {
  alerts: Alert[]
  loading: boolean
  error: Error | undefined
}

function useAlerts(
  serviceID: string,
  since: string,
  until: string,
  isValidRange: boolean,
): AlertsData {
  const depKey = `${serviceID}-${since}-${until}`
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | undefined>()
  const key = useRef(depKey)
  key.current = depKey
  const renderAlerts = useDeferredValue(alerts)

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
    if (!isValidRange) {
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

    const throttledSetAlerts = _.throttle(setAlerts, 1000)

    let [alerts, hasNextPage, endCursor, error] = await fetchAlerts('')
    if (key.current !== depKey) return // abort if the key has changed
    if (error) {
      setError(error)
      throttledSetAlerts.cancel()
      return
    }
    let allAlerts = alerts
    setAlerts(allAlerts)
    while (hasNextPage) {
      ;[alerts, hasNextPage, endCursor, error] = await fetchAlerts(endCursor)
      if (key.current !== depKey) return // abort if the key has changed
      if (error) {
        setError(error)
        throttledSetAlerts.cancel()
        return
      }
      allAlerts = allAlerts.concat(alerts)
      throttledSetAlerts(allAlerts)
    }

    setLoading(false)
  }, [depKey])

  useEffect(() => {
    fetch()
  }, [depKey])

  return {
    alerts: renderAlerts,
    loading,
    error,
  }
}

export default function AlertMetrics({
  serviceID,
}: AlertMetricsProps): JSX.Element {
  const now = useMemo(() => DateTime.now(), [])
  const minDate = now.minus({ days: MAX_DAY_COUNT - 1 }).startOf('day')
  const maxDate = now.endOf('day')

  const [params] = useURLParams({
    since: minDate.toFormat(DATE_FORMAT),
    until: maxDate.toFormat(DATE_FORMAT),
  })

  const since = DateTime.fromFormat(params.since, DATE_FORMAT).startOf('day')
  const until = DateTime.fromFormat(params.until, DATE_FORMAT).endOf('day')

  const isValidRange =
    since >= minDate &&
    until >= minDate &&
    since <= maxDate &&
    until <= maxDate &&
    since <= until

  const alertsData = useAlerts(
    serviceID,
    since.toISO(),
    until.toISO(),
    isValidRange,
  )

  if (!isValidRange) {
    return <GenericError error='The requested date range is out-of-bounds' />
  }

  if (alertsData.error) {
    return <GenericError error={alertsData.error.message} />
  }

  const ivl = Interval.fromDateTimes(since, until)

  const graphData = ivl.splitBy({ days: 1 }).map((i) => {
    const date = i.start.toLocaleString({ month: 'short', day: 'numeric' })
    const label = i.start.toLocaleString({
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })

    const bucket = alertsData.alerts.filter((a) =>
      i.contains(DateTime.fromISO(a.metrics?.closedAt as string)),
    )

    const escalatedCount = bucket.filter((a) => a.metrics?.escalated).length

    return {
      date,
      label,
      count: bucket.length,
      nonEscalatedCount: bucket.length - escalatedCount,
      escalatedCount,

      // get average of a.metrics.timeToClose values
      avgTimeToClose: bucket.length
        ? bucket.reduce((acc, a) => {
            if (!a.metrics?.timeToClose) return acc
            const timeToClose = Duration.fromISO(a.metrics.timeToClose)
            return acc + Math.ceil(timeToClose.get('minutes'))
          }, 0) / bucket.length
        : 0,

      avgTimeToAck: bucket.length
        ? bucket.reduce((acc, a) => {
            if (!a.metrics?.timeToAck) return acc
            const timeToAck = Duration.fromISO(a.metrics.timeToAck)
            return acc + Math.ceil(timeToAck.get('minutes'))
          }, 0) / bucket.length
        : 0,
    }
  })

  const daycount = Math.floor(now.diff(since, 'days').plus({ day: 1 }).days)

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Card>
          <CardHeader
            component='h2'
            title={`Daily alert metrics over the past ${daycount} days`}
          />
          <CardContent>
            <AlertMetricsFilter now={now} />
            <AlertCountGraph data={graphData} />
            <AlertAveragesGraph data={graphData} />
            <AlertMetricsTable
              alerts={alertsData.alerts.map((a) => ({
                ...a,
                ...a.metrics,
              }))}
              loading={alertsData.loading}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
