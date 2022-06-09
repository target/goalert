import React, { useMemo, useState, useEffect, useRef } from 'react'
import { Card, CardContent, CardHeader, Grid } from '@mui/material'
import { useQuery, gql } from 'urql'
import { DateTime } from 'luxon'
import { useURLParams } from '../../actions/hooks'
import AlertMetricsFilter, {
  DATE_FORMAT,
  MAX_DAY_COUNT,
} from './AlertMetricsFilter'
import AlertCountGraph from './AlertCountGraph'
import AlertMetricsTable from './AlertMetricsTable'
import { GenericError, ObjectNotFound } from '../../error-pages'
import { Alert } from '../../../schema'

const alertsQuery = gql`
  query alerts($input: AlertSearchOptions!) {
    alerts(input: $input) {
      nodes {
        id
        alertID
        summary
        details
        status
        service {
          name
          id
        }
        createdAt
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`

const metricsQuery = gql`
  query alertmetrics($rInterval: ISORInterval!, $serviceID: ID!) {
    service(id: $serviceID) {
      id
    }
    alertMetrics(
      input: { filterByServiceID: [$serviceID], rInterval: $rInterval }
    ) {
      alertCount
      timestamp
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
  const key = `${serviceID}-${since}-${until}`
  const [cursor, setCursor] = useState<string>('')
  const alertData = useRef<Record<string, Alert[]>>({})
  const dataKey = useRef<string>(key)

  if (key !== dataKey.current) {
    alertData.current = {}
    dataKey.current = key
  }

  const [{ data, fetching, error }] = useQuery({
    query: alertsQuery,
    variables: {
      input: {
        filterByServiceID: [serviceID],
        first: QUERY_LIMIT,
        notCreatedBefore: since,
        createdBefore: until,
        filterByStatus: ['StatusClosed'],
        after: cursor,
      },
    },
    pause: !isValidRange,
  })
  if (data?.alerts) alertData.current[cursor] = data.alerts.nodes as Alert[]

  useEffect(() => {
    setCursor('')
  }, [key])

  useEffect(() => {
    if (!data?.alerts?.pageInfo?.hasNextPage) return
    setCursor(data.alerts.pageInfo.endCursor)
  })

  return {
    alerts: Object.values(alertData.current).flat(),
    loading: fetching || data?.alerts?.pageInfo?.hasNextPage,
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

  const [q] = useQuery({
    query: metricsQuery,
    variables: {
      serviceID,
      rInterval: `R${Math.floor(
        until.diff(since, 'days').days,
      )}/${since.toISO()}/P1D`,
    },
    pause: !isValidRange,
  })

  if (!isValidRange) {
    return <GenericError error='The requested date range is out-of-bounds' />
  }

  if (q.error) {
    return <GenericError error={q.error.message} />
  }
  if (alertsData.error) {
    return <GenericError error={alertsData.error.message} />
  }
  if (!q.fetching && !q.data?.service?.id) {
    return <ObjectNotFound type='service' />
  }

  const alertMetrics = q.data?.alertMetrics ?? []
  const graphData = alertMetrics.map(
    (day: { timestamp: string; alertCount: number }) => {
      const timestamp = DateTime.fromISO(day.timestamp)
      const date = timestamp.toLocaleString({
        month: 'short',
        day: 'numeric',
      })
      const label = timestamp.toLocaleString({
        month: 'short',
        day: 'numeric',
        year: 'numeric',
      })
      return {
        date: date,
        count: day.alertCount,
        label: label,
      }
    },
  )

  const daycount = Math.floor(now.diff(since, 'days').plus({ day: 1 }).days)

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Card>
          <CardHeader
            component='h2'
            title={`Daily alert counts over the past ${daycount} days`}
          />
          <CardContent>
            <AlertMetricsFilter now={now} />
            <AlertCountGraph data={graphData} />
            <AlertMetricsTable
              alerts={alertsData.alerts}
              loading={alertsData.loading}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
