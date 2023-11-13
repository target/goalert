import React, { useMemo } from 'react'
import { Card, CardContent, CardHeader, Grid } from '@mui/material'
import { DateTime, DateTimeUnit, Duration, Interval } from 'luxon'
import { useURLParam } from '../../actions/hooks'
import AlertMetricsFilter from './AlertMetricsFilter'
import AlertCountGraph from './AlertCountGraph'
import AlertMetricsTable from './AlertMetricsTable'
import AlertAveragesGraph from './AlertAveragesGraph'
import { GenericError, ObjectNotFound } from '../../error-pages'
import { useWorker } from '../../worker'
import { AlertMetricsOpts } from './useAlertMetrics'
import { useAlerts } from './useAlerts'
import { useQuery } from 'urql'
import Spinner from '../../loading/components/Spinner'
import { AlertSearchOptions } from '../../../schema'

export type AlertMetricsProps = {
  serviceID: string
}

const units: Record<string, DateTimeUnit> = {
  P1D: 'day',
  P1W: 'week',
  P1M: 'month',
}

export default function AlertMetrics({
  serviceID,
}: AlertMetricsProps): React.ReactNode {
  const now = useMemo(() => DateTime.now(), [])

  const [svc] = useQuery({
    query: 'query Svc($id: ID!) {service(id:$id){id,name}}',
    variables: { id: serviceID },
  })
  const [range] = useURLParam('range', 'P1M')
  const [ivl] = useURLParam('interval', 'P1D')
  const graphDur = Duration.fromISO(ivl).toISO()

  const unit = units[ivl]
  const since = now.minus(Duration.fromISO(range)).startOf(unit)
  const until = now

  const alertOptions: AlertSearchOptions = {
    filterByServiceID: [serviceID],
    filterByStatus: ['StatusClosed'],
    notClosedBefore: since.toISO(),
    closedBefore: until.toISO(),
  }
  const depKey = `${serviceID}-${since}-${until}`

  const alertsData = useAlerts(alertOptions, depKey)
  const graphInterval = Interval.fromDateTimes(since, until).toISO()

  // useMemo to use same object reference
  const metricsOpts: AlertMetricsOpts = useMemo(
    () => ({ int: graphInterval, dur: graphDur, alerts: alertsData.alerts }),
    [graphInterval, graphDur, alertsData.alerts],
  )

  const [graphData, graphDataStatus] = useWorker(
    'useAlertMetrics',
    metricsOpts,
    [],
  )

  if (svc.fetching) return <Spinner />
  if (!svc.data?.service?.name) return <ObjectNotFound />

  if (alertsData.error) {
    return <GenericError error={alertsData.error.message} />
  }

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Card>
          <CardContent>
            <AlertMetricsFilter />
            <CardHeader
              title='Alert Counts'
              component='h2'
              sx={{ ml: '2rem', mb: 0, pb: 0 }}
            />
            <AlertCountGraph
              data={graphData}
              loading={graphDataStatus.loading || alertsData.loading}
            />
            <CardHeader
              title='Alert Averages'
              component='h2'
              sx={{ ml: '2rem', mb: 0, pb: 0 }}
            />
            <AlertAveragesGraph
              data={graphData}
              loading={graphDataStatus.loading || alertsData.loading}
            />
            <AlertMetricsTable
              alerts={alertsData.alerts}
              serviceName={svc.data.service.name}
              startTime={since.toFormat('yyyy-MM-dd')}
              endTime={until.toFormat('yyyy-MM-dd')}
              loading={graphDataStatus.loading || alertsData.loading}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
