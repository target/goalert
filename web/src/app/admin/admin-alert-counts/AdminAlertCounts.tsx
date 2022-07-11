import React, { useMemo } from 'react'
import { Grid, Card, CardContent } from '@mui/material'
import AlertCountControls from './AlertCountControls'
import { useURLParams } from '../../actions'
import { DateTime, Duration, Interval } from 'luxon'
import { AlertSearchOptions } from '../../../schema'
import { useAlerts } from '../../services/AlertMetrics/useAlerts'
import { GenericError } from '../../error-pages'
import { useWorker } from '../../worker'
import AlertCountLineGraph from './AlertCountLineGraph'
import AlertCountTable from './AlertCountTable'
import { AlertCountOpts } from './useAdminAlertCounts'

export default function AdminAlertCounts(): JSX.Element {
  const now = useMemo(() => DateTime.now(), [])
  const [params] = useURLParams({
    since: now.minus({ months: 1 }).toISO(),
    until: now.toISO(),
    interval: 'P1D',
  })

  const graphDur = Duration.fromISO(params.interval).toISO()
  const graphInterval = Interval.fromDateTimes(
    DateTime.fromISO(params.since),
    DateTime.fromISO(params.until),
  ).toISO()

  const alertOptions: AlertSearchOptions = {
    createdBefore: params.until,
    notCreatedBefore: params.since,
  }
  const depKey = `${params.since}-${params.until}`
  const alertsData = useAlerts(alertOptions, depKey)

  const alertCountOpts: AlertCountOpts = useMemo(
    () => ({ int: graphInterval, dur: graphDur, alerts: alertsData.alerts }),
    [graphInterval, graphDur, alertsData.alerts],
  )
  const alertCounts = useWorker('useAdminAlertCounts', alertCountOpts, [])

  if (alertsData.error) {
    return <GenericError error={alertsData.error.message} />
  }

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <AlertCountControls />
        <Card sx={{ marginTop: '15px' }}>
          <CardContent>
            <AlertCountLineGraph data={alertCounts.slice(0, 5)} />
            <AlertCountTable
              alertCounts={alertCounts}
              startTime={DateTime.fromISO(params.since).toFormat('yyyy-MM-dd')}
              endTime={DateTime.fromISO(params.until).toFormat('yyyy-MM-dd')}
              loading={alertsData.loading}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
