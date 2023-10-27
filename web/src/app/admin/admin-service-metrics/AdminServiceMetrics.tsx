import React, { useMemo } from 'react'
import { Grid, Card, CardHeader, CardContent, Tooltip } from '@mui/material'
import { DateTime } from 'luxon'
import { useServices } from './useServices'
import { useWorker } from '../../worker'
import { ServiceMetrics } from './useServiceMetrics'
import AdminServiceTable from './AdminServiceTable'
import {
  ErrorOutline,
  WarningAmberOutlined,
  NotificationsOffOutlined,
  UpdateDisabledOutlined,
} from '@mui/icons-material'
import { AlertSearchOptions, Service } from '../../../schema'
import { useAlerts } from '../../services/AlertMetrics/useAlerts'
import AdminServiceFilter from './AdminServiceFilter'
import { useURLParams } from '../../actions'

const STALE_ALERT_LIMIT = 2

export default function AdminServiceMetrics(): JSX.Element {
  const now = useMemo(() => DateTime.now(), [])
  const [params] = useURLParams({
    epStepTgts: [] as string[],
    intKeyTgts: [] as string[],
    labelKey: '',
    labelValue: '',
  })

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
      filters: {
        labelKey: params.labelKey,
        labelValue: params.labelValue,
        epStepTgts: params.epStepTgts,
        intKeyTgts: params.intKeyTgts,
      },
    },
    {} as ServiceMetrics,
  )

  const getConfigIssueCounts = (
    services: Service[],
  ): {
    totalNoIntegration: number
    totalNoEP: number
    totalAlertLimit: number
  } => {
    return services.reduce(
      (counts, svc) => {
        if (!svc.integrationKeys.length) {
          if (!svc.heartbeatMonitors.length) counts.totalNoIntegration++
        }
        if (!svc.escalationPolicy?.steps.length) counts.totalNoEP++
        else if (
          svc.escalationPolicy.steps?.every((step) => !step.targets.length)
        )
          counts.totalNoEP++
        if (svc.notices.length) counts.totalAlertLimit++
        return counts
      },
      {
        totalNoIntegration: 0,
        totalNoEP: 0,
        totalAlertLimit: 0,
      },
    )
  }

  const { totalNoIntegration, totalNoEP, totalAlertLimit } =
    getConfigIssueCounts(serviceData.services || [])

  const cardSubHeader = serviceData.loading
    ? 'Loading services... This may take a minute'
    : `Metrics pulled from ${metrics.filteredServices.length} services`

  return (
    <Grid container spacing={2}>
      <Grid item xs={4} sm={2.4}>
        <Card sx={{ height: '100%' }}>
          <CardHeader
            title={serviceData.services.length}
            subheader='Total Services'
          />
        </Card>
      </Grid>
      <Grid item xs={4} sm={2.4}>
        <Card sx={{ height: '100%' }}>
          <CardHeader
            title={totalNoIntegration}
            subheader='Services With No Integrations'
            action={
              !!totalNoIntegration && (
                <Tooltip title='Services with no integration keys or heartbeat monitors.'>
                  <WarningAmberOutlined color='warning' />
                </Tooltip>
              )
            }
          />
        </Card>
      </Grid>
      <Grid item xs={4} sm={2.4}>
        <Card sx={{ height: '100%' }}>
          <CardHeader
            title={totalNoEP}
            subheader='Services With Empty Escalation Policies'
            action={
              !!totalNoEP && (
                <Tooltip title='Services with empty escalation policies.'>
                  <NotificationsOffOutlined color='error' />
                </Tooltip>
              )
            }
          />
        </Card>
      </Grid>
      <Grid item xs={4} sm={2.4}>
        <Card sx={{ height: '100%' }}>
          <CardHeader
            title={
              metrics.totalStaleAlerts
                ? Object.keys(metrics.totalStaleAlerts).length
                : 0
            }
            subheader='Services With Stale Alerts'
            action={
              !!metrics.totalStaleAlerts && (
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
      <Grid item xs={3} sm={2.4}>
        <Card sx={{ height: '100%' }}>
          <CardHeader
            title={totalAlertLimit}
            subheader='Services Reaching Alert Limit'
            action={
              !!totalAlertLimit && (
                <Tooltip title='Services at or nearing unacknowledged alert limit.'>
                  <ErrorOutline color='error' />
                </Tooltip>
              )
            }
          />
        </Card>
      </Grid>
      <Grid item xs={12}>
        <AdminServiceFilter />
      </Grid>
      <Grid item xs={12}>
        <Card sx={{ marginTop: (theme) => theme.spacing(1) }}>
          <CardHeader title='Services' subheader={cardSubHeader} />
          <CardContent>
            <AdminServiceTable
              services={metrics.filteredServices}
              staleAlertServices={metrics.totalStaleAlerts}
              loading={serviceData.loading}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
