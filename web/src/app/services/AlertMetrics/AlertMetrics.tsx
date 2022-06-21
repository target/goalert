import React, { useMemo } from 'react'
import { Card, CardContent, CardHeader, Grid } from '@mui/material'
import { DateTime, Duration, Interval } from 'luxon'
import { useURLParam, useURLParams } from '../../actions/hooks'
import AlertMetricsFilter, {
  DATE_FORMAT,
  MAX_DAY_COUNT,
} from './AlertMetricsFilter'
import AlertCountGraph from './AlertCountGraph'
import AlertMetricsTable from './AlertMetricsTable'
import AlertAveragesGraph from './AlertAveragesGraph'
import { GenericError, ObjectNotFound } from '../../error-pages'
import _ from 'lodash'
import { useWorker } from '../../worker'
import { AlertMetricsOpts, useAlertMetrics } from './useAlertMetrics'
import { useAlerts } from './useAlerts'
import { useQuery } from 'urql'
import Spinner from '../../loading/components/Spinner'

export type AlertMetricsProps = {
  serviceID: string
}

export default function AlertMetrics({
  serviceID,
}: AlertMetricsProps): JSX.Element {
  const now = useMemo(() => DateTime.now(), [])

  const [svc] = useQuery({
    query: 'query Svc($id: ID!) {service(id:$id){id,name}}',
    variables: { id: serviceID },
  })
  const [range] = useURLParam('range', 'P1M')
  const [ivl] = useURLParam('interval', 'P1D')

  const since = now.minus(Duration.fromISO(range)).startOf('day')
  const until = now.endOf('day')

  const alertsData = useAlerts(serviceID, since.toISO(), until.toISO(), true)
  const graphInterval = Interval.fromDateTimes(since, until).toISO()
  const graphDur = Duration.fromISO(ivl).toISO()

  // useMemo to use same object reference
  const metricsOpts: AlertMetricsOpts = useMemo(
    () => ({ int: graphInterval, dur: graphDur, alerts: alertsData.alerts }),
    [graphInterval, graphDur, alertsData.alerts],
  )

  const graphData = useWorker(useAlertMetrics, metricsOpts, [])

  if (svc.fetching) return <Spinner />
  if (!svc.data?.service?.name) return <ObjectNotFound />

  if (alertsData.error) {
    return <GenericError error={alertsData.error.message} />
  }

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
              alerts={alertsData.alerts}
              serviceName={svc.data.service.name}
              startTime={since.toFormat('yyyy-MM-dd')}
              endTime={until.toFormat('yyyy-MM-dd')}
              loading={alertsData.loading}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
