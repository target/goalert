import React, { useMemo } from 'react'
import { Card, CardContent, CardHeader, Grid } from '@mui/material'
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

const query = gql`
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

interface AlertMetricsProps {
  serviceID: string
}

export default function AlertMetrics({
  serviceID,
}: AlertMetricsProps): JSX.Element {
  const now = useMemo(() => DateTime.now(), [])
  const minDate = useMemo(
    () => now.minus({ weeks: MAX_WEEKS_COUNT }).startOf('day'),
    [now],
  )

  const [_since] = useURLParam('since', minDate.toFormat(DATE_FORMAT))
  const since = DateTime.max(DateTime.fromFormat(_since, DATE_FORMAT), minDate) // set a floor

  const q = useQuery(query, {
    variables: {
      input: {
        filterByServiceID: [serviceID],
        first: 100,
        notCreatedBefore: since.toISO(),
        createdBefore: now.toISO(),
      },
    },
  })

  const alerts = q?.data?.alerts?.nodes ?? []

  const dateToAlerts = _.groupBy(alerts, (node) =>
    DateTime.fromISO(node.createdAt).toLocaleString({
      month: 'short',
      day: 'numeric',
    }),
  )

  const data = Interval.fromDateTimes(
    since.plus({ day: 1 }).startOf('day'),
    now.endOf('day'),
  )
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

  const daycount = Math.floor(now.diff(since, 'days').days)

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
            <AlertCountGraph data={data} />
            <AlertMetricsTable alerts={alerts} />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
