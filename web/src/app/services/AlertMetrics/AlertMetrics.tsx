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
}: AlertMetricsProps): JSX.Element {
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

  function getRange(range: string): string {
    switch (range) {
      case 'P1W':
        return 'past week'
      case 'P2W':
        return 'past 2 weeks'
      case 'P1M':
        return 'past month'
      case 'P3M':
        return 'past 3 months'
      case 'P6M':
        return 'past 6 months'
      case 'P1Y':
        return 'past year'
      default:
        return `past ${Math.ceil(until.diff(since, unit).as(unit))} ${unit}s`
    }
  }

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Card>
          <CardHeader
            component='h2'
            title={`Daily alert metrics over the ${getRange(range)}`}
          />
          <CardContent>
            <AlertMetricsFilter />
            <AlertCountGraph
              data={graphData}
              loading={graphDataStatus.loading || alertsData.loading}
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
