import React, { useMemo } from 'react'
import { Card, CardContent, CardHeader, Grid } from '@mui/material'
import { DateTime, Duration, Interval } from 'luxon'
import { useURLParams } from '../../actions/hooks'
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
  const minDate = now.minus({ days: MAX_DAY_COUNT - 1 }).startOf('day')
  const maxDate = now.endOf('day')

  const [svc] = useQuery({
    query: 'query Svc($id: ID!) {service(id:$id){id,name}}',
    variables: { id: serviceID },
  })

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
  const graphInterval = Interval.fromDateTimes(since, until).toISO()
  const graphDur = Duration.fromObject({ days: 1 }).toISO()

  // useMemo to use same object reference
  const metricsOpts: AlertMetricsOpts = useMemo(
    () => ({ int: graphInterval, dur: graphDur, alerts: alertsData.alerts }),
    [graphInterval, graphDur, alertsData.alerts],
  )

  const useAlertMetricsFn = useWorker(useAlertMetrics)
  const graphData = useAlertMetricsFn(metricsOpts)

  if (svc.fetching) return <Spinner />
  if (!svc.data?.service?.name) return <ObjectNotFound />
  if (!isValidRange) {
    return <GenericError error='The requested date range is out-of-bounds' />
  }

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
