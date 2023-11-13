import React, { useMemo, useState } from 'react'
import { Grid, Card, CardContent, CardHeader } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import AlertCountControls from './AlertCountControls'
import { useURLParams } from '../../actions'
import { DateTime, Duration, Interval, DateTimeUnit } from 'luxon'
import { AlertSearchOptions } from '../../../schema'
import { useAlerts } from '../../services/AlertMetrics/useAlerts'
import { GenericError } from '../../error-pages'
import { useWorker } from '../../worker'
import AlertCountLineGraph from './AlertCountLineGraph'
import AlertCountTable from './AlertCountTable'
import { AlertCountOpts, AlertCountSeries } from './useAdminAlertCounts'

const useStyles = makeStyles((theme: Theme) => ({
  card: {
    marginTop: theme.spacing(1),
  },
}))

const units: Record<string, DateTimeUnit> = {
  PT1M: 'minute',
  PT1H: 'hour',
  P1D: 'day',
  P1W: 'week',
  P1M: 'month',
}

export default function AdminAlertCounts(): React.ReactNode {
  const styles = useStyles()

  const [graphData, setGraphData] = useState<AlertCountSeries[]>([])
  const now = useMemo(() => DateTime.now(), [])

  const [params] = useURLParams({
    createdAfter: now.minus({ days: 1 }).toISO(),
    createdBefore: '',
    interval: 'PT1H',
  })

  const unit = units[params.interval]
  const until = params.createdBefore
    ? DateTime.fromISO(params.createdBefore)
    : now.startOf(unit)

  const graphDur = Duration.fromISO(params.interval).toISO()
  const graphInterval = Interval.fromDateTimes(
    DateTime.fromISO(params.createdAfter),
    until,
  ).toISO()

  const alertOptions: AlertSearchOptions = {
    notCreatedBefore: params.createdAfter,
  }
  if (params.createdBefore) {
    alertOptions.createdBefore = params.createdBefore
  }
  const depKey = `${params.createdAfter}-${until.toISO()}`
  const alertsData = useAlerts(alertOptions, depKey)

  const alertCountOpts: AlertCountOpts = useMemo(
    () => ({ int: graphInterval, dur: graphDur, alerts: alertsData.alerts }),
    [graphInterval, graphDur, alertsData.alerts],
  )
  const [alertCounts, alertCountStatus] = useWorker(
    'useAdminAlertCounts',
    alertCountOpts,
    [],
  )

  if (alertsData.error) {
    return <GenericError error={alertsData.error.message} />
  }

  const dayCount = Math.ceil(
    until.diff(DateTime.fromISO(params.createdAfter), unit).as(unit),
  )

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <AlertCountControls />
        <Card className={styles.card}>
          <CardHeader
            title='Alert Counts'
            subheader={`Number of alerts created by services over the past ${dayCount} ${unit}s`}
          />
          <CardContent>
            <AlertCountLineGraph
              data={graphData}
              loading={alertCountStatus.loading || alertsData.loading}
              unit={unit}
            />
            <AlertCountTable
              alertCounts={alertCounts}
              graphData={graphData}
              setGraphData={setGraphData}
              startTime={DateTime.fromISO(params.createdAfter).toFormat(
                'yyyy-MM-dd',
              )}
              endTime={until.toFormat('yyyy-MM-dd')}
              loading={alertCountStatus.loading || alertsData.loading}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
