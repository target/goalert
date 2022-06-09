import React, { useMemo, useState, useEffect } from 'react'
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
        hasNextPage
        endCursor
      }
    }
    alertMetrics(input: $alertMetricsInput) {
      alertCount
      timestamp
    }
  }
`

const QUERY_LIMIT = 100

export type AlertMetricsProps = {
  serviceID: string
}

export default function AlertMetrics({
  serviceID,
}: AlertMetricsProps): JSX.Element {
  const [alertsList, setAlertsList] = useState<Alert[]>([])
  const [endCursor, setEndCursor] = useState()

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

  const [q] = useQuery({
    query,
    variables: {
      serviceID,
      alertSearchInput: {
        filterByServiceID: [serviceID],
        first: QUERY_LIMIT,
        notCreatedBefore: since.toISO(),
        createdBefore: until.toISO(),
        filterByStatus: ['StatusClosed'],
        after: endCursor,
      },
      alertMetricsInput: {
        rInterval: `R${Math.floor(
          until.diff(since, 'days').days,
        )}/${since.toISO()}/P1D`,
        filterByServiceID: [serviceID],
      },
    },
    pause: !isValidRange,
  })

  useEffect(() => {
    if (q.data) {
      for (let i = 0; i < q.data?.alerts?.nodes.length; i++) {
        // Do not save duplciate alerts to state
        if (
          !(
            alertsList.filter(
              (alert) => alert.id === q.data?.alerts?.nodes[i].id,
            ).length > 0
          )
        ) {
          setAlertsList((prev) => [...prev, q.data?.alerts?.nodes[i]])
        }
      }

      // Update endCursor if hasNextPage
      if (q.data?.alerts?.pageInfo?.hasNextPage) {
        setEndCursor(q.data?.alerts?.pageInfo?.endCursor)
      }
    }
  }, [q])

  if (!isValidRange) {
    return <GenericError error='The requested date range is out-of-bounds' />
  }

  if (q.error) {
    return <GenericError error={q.error.message} />
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
              alerts={alertsList.filter(function (alert) {
                return DateTime.fromISO(alert.createdAt) >= since
              })}
              loading={q.fetching || !alertsList}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
