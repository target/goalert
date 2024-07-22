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
import Spinner from '../../loading/components/Spinner'
import { useURLParams } from '../../actions'
import AdminServiceTargetGraph from './AdminServiceTargetGraph'

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
    totalNoIntKey: number
    totalAlertLimit: number
  } => {
    return services.reduce(
      (counts, svc) => {
        if (!svc.integrationKeys.length) {
          if (!svc.heartbeatMonitors.length) counts.totalNoIntegration++
          counts.totalNoIntKey++
        }
        if (!svc.escalationPolicy?.steps.length) counts.totalNoEP++
        else if (
          svc.escalationPolicy.steps?.every((step) => !step.actions.length)
        )
          counts.totalNoEP++
        if (svc.notices.length) counts.totalAlertLimit++
        return counts
      },
      {
        totalNoIntegration: 0,
        totalNoEP: 0,
        totalNoIntKey: 0,
        totalAlertLimit: 0,
      },
    )
  }

  const { totalNoIntegration, totalNoEP, totalAlertLimit } =
    getConfigIssueCounts(serviceData.services || [])

  const { totalNoIntKey: filteredTotalNoIntKey, totalNoEP: filteredTotalNoEP } =
    getConfigIssueCounts(metrics.filteredServices || [])

  const cardSubHeader = serviceData.loading
    ? 'Loading services... This may take a minute'
    : `Metrics pulled from ${metrics.filteredServices.length} services`

  function renderOverviewMetrics(): JSX.Element {
    return (
      <React.Fragment>
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
      </React.Fragment>
    )
  }

  function renderUsageGraphs(): JSX.Element {
    return (
      <React.Fragment>
        <Grid item xs>
          <Card sx={{ marginTop: (theme) => theme.spacing(1) }}>
            <CardHeader
              title='Integration Key Usage'
              subheader={
                metrics.filteredServices?.length +
                ' services, of which ' +
                filteredTotalNoIntKey +
                ' service(s) have no integration keys configured.'
              }
            />
            <CardContent>
              <AdminServiceTargetGraph
                metrics={metrics.keyTgtTotals}
                loading={serviceData.loading}
              />
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs>
          <Card sx={{ marginTop: (theme) => theme.spacing(1) }}>
            <CardHeader
              title='Escalation Policy Usage'
              subheader={
                metrics.filteredServices?.length +
                ' services, of which ' +
                filteredTotalNoEP +
                ' service(s) have empty escalation policies.'
              }
            />
            <CardContent>
              <AdminServiceTargetGraph
                metrics={metrics.stepTgtTotals}
                loading={serviceData.loading}
              />
            </CardContent>
          </Card>
        </Grid>
      </React.Fragment>
    )
  }

  function renderServiceTable(): JSX.Element {
    return (
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
    )
  }

  return (
    <Grid container spacing={2}>
      {serviceData.loading && <Spinner />}
      {renderOverviewMetrics()}
      <Grid item xs={12}>
        <AdminServiceFilter />
      </Grid>
      {renderUsageGraphs()}
      <Grid item xs={12}>
        {renderServiceTable()}
      </Grid>
    </Grid>
  )
}
