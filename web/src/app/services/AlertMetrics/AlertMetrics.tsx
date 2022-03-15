import React, { useMemo, useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, Grid } from '@mui/material'
import { useQuery, gql } from '@apollo/client'
import { DateTime } from 'luxon'
import { useParams } from 'react-router-dom'
import { useURLParams } from '../../actions/hooks'
import AlertMetricsFilter, {
  DATE_FORMAT,
  MAX_DAY_COUNT,
} from './AlertMetricsFilter'
import AlertCountGraph from './AlertCountGraph'
import AlertMetricsTable from './AlertMetricsTable'
import { GenericError, ObjectNotFound } from '../../error-pages'
import { POLL_INTERVAL } from '../../config'
import { Alert } from '../../../schema'

const query = gql`
  query alertmetrics(
    $serviceID: ID!
    $alertSearchInput: AlertSearchOptions!
    $alertMetricsInput: AlertMetricsOptions!
  ) {
    service(id: $serviceID) {
      id
    }
    alerts(input: $alertSearchInput) {
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
        endCursor
        hasNextPage
      }
    }
    alertMetrics(input: $alertMetricsInput) {
      alertCount
      timestamp
    }
  }
`

const QUERY_LIMIT = 100

export default function AlertMetrics(): JSX.Element {
  const [poll, setPoll] = useState(POLL_INTERVAL)
  const { serviceID } = useParams<{ serviceID: string }>()
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

  const q = useQuery(query, {
    pollInterval: poll,
    variables: {
      serviceID,
      alertSearchInput: {
        filterByServiceID: [serviceID],
        first: QUERY_LIMIT,
        notCreatedBefore: since.toISO(),
        createdBefore: until.toISO(),
        filterByStatus: ['StatusClosed'],
      },
      alertMetricsInput: {
        rInterval: `R${Math.floor(
          until.diff(since, 'days').days,
        )}/${since.toISO()}/P1D`,
        filterByServiceID: [serviceID],
      },
    },
    skip: !isValidRange,
  })

  useEffect(() => {
    if (!q.loading && q?.data?.alerts?.pageInfo?.hasNextPage) {
      setPoll(0)
      q.fetchMore({
        variables: {
          input: {
            first: QUERY_LIMIT,
            after: q.data?.alerts?.pageInfo?.endCursor,
          },
        },
        updateQuery: (
          prev: { alerts: { nodes: Alert[] } },
          { fetchMoreResult },
        ) => {
          if (!fetchMoreResult) return prev
          let alerts = prev.alerts.nodes.concat(fetchMoreResult.alerts.nodes)
          alerts = alerts.filter((alert, idx) => alerts.indexOf(alert) === idx)
          return {
            alerts: {
              ...fetchMoreResult.alerts,
              nodes: alerts,
            },
          }
        },
      })
    }
  }, [q])

  if (!isValidRange) {
    return <GenericError error='The requested date range is out-of-bounds' />
  }

  if (q.error) {
    return <GenericError error={q.error.message} />
  }
  if (!q.loading && !q.data?.service?.id) {
    return <ObjectNotFound type='service' />
  }

  const hasNextPage = q.data?.alerts?.pageInfo?.hasNextPage ?? false
  const alerts = q.data?.alerts?.nodes ?? []
  const alertMetrics = q.data?.alertMetrics ?? []

  const data = alertMetrics.map(
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
            <AlertCountGraph
              data={data}
              loading={q.loading || !q?.data?.alerts || hasNextPage}
            />
            <AlertMetricsTable
              alerts={alerts}
              loading={q.loading || !q?.data?.alerts || hasNextPage}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
