import React, { useMemo } from 'react'
import { Grid, Card, CardHeader, CardContent, Tooltip } from '@mui/material'
import { DateTime } from 'luxon'
import { useServices } from './useServices'
import { useWorker } from '../../worker'
import { ServiceMetrics } from './useServiceMetrics'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import AdminServiceTable from './AdminServiceTable'
import {
  ErrorOutline,
  WarningAmberOutlined,
  NotificationsOffOutlined,
  UpdateDisabledOutlined,
} from '@mui/icons-material'
import { AlertSearchOptions, Service } from '../../../schema'
import { useAlerts } from '../../services/AlertMetrics/useAlerts'

const useStyles = makeStyles((theme: Theme) => ({
  card: {
    marginTop: theme.spacing(1),
  },
}))

const STALE_ALERT_LIMIT = 2

export default function AdminServiceMetrics(): JSX.Element {
  const now = useMemo(() => DateTime.now(), [])
  const styles = useStyles()

  const depKey = `${now}`
  const serviceData = useServices(depKey)

  const alertOptions: AlertSearchOptions = {
    createdBefore: now.minus({ month: STALE_ALERT_LIMIT }).toISO(),
    filterByStatus: ['StatusAcknowledged'],
  }

  const alertsData = useAlerts(alertOptions, depKey)

  const [metrics] = useWorker(
    'useServiceMetrics',
    {
      services: serviceData.services,
      alerts: alertsData.alerts,
    },
    {} as ServiceMetrics,
  )

  const getConfigIssueCounts = (
    services: Service[],
  ): {
    noIntegrationTotal: number
    noEPTotal: number
    alertLimitTotal: number
  } => {
    let noIntegrationTotal = 0
    let noEPTotal = 0
    let alertLimitTotal = 0

    services.map((svc: Service) => {
      if (!svc.heartbeatMonitors.length && !svc.integrationKeys.length)
        noIntegrationTotal += 1
      if (!svc.escalationPolicy?.steps.length) noEPTotal += 1
      else if (
        svc.escalationPolicy.steps.every((step) => step.targets.length === 0)
      ) {
        noEPTotal += 1
      }
      if (svc.notices.length) alertLimitTotal += 1
    })
    return { noIntegrationTotal, noEPTotal, alertLimitTotal }
  }

  const { noIntegrationTotal, noEPTotal, alertLimitTotal } =
    getConfigIssueCounts(serviceData.services || [])

  return (
    <Grid container spacing={2}>
      <Grid item xs={2.4}>
        <Card>
          <CardHeader
            title={serviceData.services.length}
            subheader='Total Services'
          />
        </Card>
      </Grid>
      <Grid item xs={2.4}>
        <Card>
          <CardHeader
            title={noIntegrationTotal}
            subheader='Services With No Integrations'
            action={
              !!noIntegrationTotal && (
                <Tooltip title='Services with no integration keys or heartbeat monitors.'>
                  <WarningAmberOutlined color='warning' />
                </Tooltip>
              )
            }
          />
        </Card>
      </Grid>
      <Grid item xs={2.4}>
        <Card>
          <CardHeader
            title={noEPTotal}
            subheader='Services With Empty Escalation Policies'
            action={
              !!noEPTotal && (
                <Tooltip title='Services with empty escalation policies.'>
                  <NotificationsOffOutlined color='error' />
                </Tooltip>
              )
            }
          />
        </Card>
      </Grid>
      <Grid item xs={2.4}>
        <Card>
          <CardHeader
            title={
              metrics.staleAlertTotals
                ? Object.keys(metrics.staleAlertTotals).length
                : 0
            }
            subheader='Services With Stale Alerts'
            action={
              !!metrics.staleAlertTotals && (
                <Tooltip
                  title={`Services with acknowledged alerts created more than ${STALE_ALERT_LIMIT} months ago.`}
                >
                  <UpdateDisabledOutlined color='warning' />
                </Tooltip>
              )
            }
          />
        </Card>
      </Grid>
      <Grid item xs={2.4}>
        <Card>
          <CardHeader
            title={alertLimitTotal}
            subheader='Services Reaching Alert Limit'
            action={
              !!alertLimitTotal && (
                <Tooltip title='Services at or nearing unacknowledged alert limit.'>
                  <ErrorOutline color='error' />
                </Tooltip>
              )
            }
          />
        </Card>
      </Grid>
      <Grid item xs={12}>
        <Card className={styles.card}>
          <CardHeader title='Services' />
          <CardContent>
            <AdminServiceTable
              services={metrics.filteredServices}
              staleAlertServices={metrics.staleAlertTotals}
              loading={serviceData.loading}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
