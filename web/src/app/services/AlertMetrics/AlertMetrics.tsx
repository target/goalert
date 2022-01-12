import React, { useMemo } from 'react'
import { Box, Card, CardContent, CardHeader, Grid } from '@mui/material'
import { useQuery, gql } from '@apollo/client'
import { DateTime, Interval } from 'luxon'
import _ from 'lodash'
import { useURLParam } from '../../actions/hooks'
import AlertMetricsFilter, {
  DATE_FORMAT,
  MAX_WEEKS_COUNT,
} from './AlertMetricsFilter'
import AlertCountGraph from './AlertCountGraph'
import AlertMetricsTable from './AlertMetricsTable'
import Notices from '../../details/Notices'
import { GenericError, ObjectNotFound } from '../../error-pages'

const query = gql`
  query alertmetrics($serviceID: ID!, $input: AlertSearchOptions!) {
    service(id: $serviceID) {
      id
    }
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
      }
    }
  }
`

interface AlertMetricsProps {
  serviceID: string
}

const QUERY_LIMIT = 100

export default function AlertMetrics({
  serviceID,
}: AlertMetricsProps): JSX.Element {
  const now = useMemo(() => DateTime.now(), [])
  const [minTime, maxTime] = [
    now.minus({ weeks: MAX_WEEKS_COUNT }).plus({ days: 1 }).startOf('day'),
    now,
  ]

  const [_since] = useURLParam('since', minTime.toFormat(DATE_FORMAT))
  const since = DateTime.fromFormat(_since, DATE_FORMAT)

  const isValidRange = since >= minTime && since < maxTime

  const q = useQuery(query, {
    variables: {
      serviceID,
      input: {
        filterByServiceID: [serviceID],
        first: QUERY_LIMIT,
        notCreatedBefore: since.toISO(),
        createdBefore: now.toISO(),
      },
    },
    skip: !isValidRange,
  })

  if (!isValidRange) {
    return <GenericError error='The requested date range is out-of-bounds' />
  }

  if (q.error) {
    return <GenericError error={q.error.message} />
  }
  if (!q.loading && !q.data?.service?.id) {
    return <ObjectNotFound type='service' />
  }

  const hasNextPage = q?.data?.alerts?.pageInfo?.hasNextPage ?? false
  const alerts = q?.data?.alerts?.nodes ?? []

  const dateToAlerts = _.groupBy(alerts, (node) =>
    DateTime.fromISO(node.createdAt).toLocaleString({
      month: 'short',
      day: 'numeric',
    }),
  )

  const data = Interval.fromDateTimes(since.startOf('day'), now.endOf('day'))
    .splitBy({ days: 1 })
    .map((day) => {
      let alertCount = 0
      const date = day.start.toLocaleString({ month: 'short', day: 'numeric' })

      if (dateToAlerts[date]) {
        alertCount = dateToAlerts[date].length
      }

      return {
        date: date,
        count: alertCount,
      }
    })

  const daycount = Math.floor(now.diff(since, 'days').plus({ day: 1 }).days)

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        {hasNextPage && (
          <Box sx={{ marginBottom: '1rem' }}>
            <Notices
              notices={[
                {
                  type: 'WARNING',
                  message: 'Query limit reached',
                  details: `More than ${QUERY_LIMIT} alerts were found, but only the first ${QUERY_LIMIT} are represented below.`,
                },
              ]}
            />
          </Box>
        )}
        <Card>
          <CardHeader
            component='h2'
            title={`Daily alert counts over the past ${daycount} days`}
          />
          <CardContent>
            <AlertMetricsFilter now={now} />
            <AlertCountGraph data={data} />
            <AlertMetricsTable
              alerts={alerts}
              loading={q.loading || !q?.data?.alerts}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
