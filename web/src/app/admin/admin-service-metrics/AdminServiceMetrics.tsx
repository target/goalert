import React, { useMemo } from 'react'
import { Grid, Card, CardHeader, CardContent, Tooltip } from '@mui/material'
import { DateTime } from 'luxon'
import { useServices } from './useServices'
import { useWorker } from '../../worker'
import { ServiceMetrics } from './useServiceMetrics'
import AdminServiceTargetGraph from './AdminServiceTargetGraph'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import AdminServiceTable from './AdminServiceTable'
import AdminServiceFilter from './AdminServiceFilter'
import { useURLParams } from '../../actions'
import { ErrorOutline, WarningAmberOutlined } from '@mui/icons-material'
import Spinner from '../../loading/components/Spinner'
import { Service } from '../../../schema'

const useStyles = makeStyles((theme: Theme) => ({
  card: {
    marginTop: theme.spacing(1),
  },
}))

export default function AdminServiceMetrics(): JSX.Element {
  const now = useMemo(() => DateTime.now(), [])
  const styles = useStyles()
  const [params] = useURLParams({
    epStepTgts: [] as string[],
    intKeyTgts: [] as string[],
    labelKey: '',
    labelValue: '',
  })

  const depKey = `${now}`
  const serviceData = useServices(depKey)
  const [metrics] = useWorker(
    'useServiceMetrics',
    {
      services: serviceData.services,
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
    noIntegrationTotal: number
    noNotifTotal: number
    noIntKeyTotal: number
    alertLimitTotal: number
  } => {
    return services.reduce(
      (counts, svc) => {
        if (!svc.integrationKeys.length) {
          if (!svc.heartbeatMonitors.length) counts.noIntegrationTotal++
          counts.noIntKeyTotal++
        }
        if (!svc.escalationPolicy?.steps.length) counts.noNotifTotal++
        else if (
          svc.escalationPolicy.steps?.every((step) => !step.targets.length)
        )
          counts.noNotifTotal++
        if (svc.notices.length) counts.alertLimitTotal++
        return counts
      },
      {
        noIntegrationTotal: 0,
        noIntKeyTotal: 0,
        noNotifTotal: 0,
        alertLimitTotal: 0,
      },
    )
  }

  const { noIntegrationTotal, noNotifTotal, alertLimitTotal } =
    getConfigIssueCounts(serviceData.services || [])

  const {
    noIntKeyTotal: filteredNoIntKeyTotal,
    noNotifTotal: filteredNoNotifTotal,
  } = getConfigIssueCounts(metrics.filteredServices || [])

  function renderOverviewMetrics(): JSX.Element {
    return (
      <React.Fragment>
        <Grid item xs>
          <Card>
            <CardHeader
              title={serviceData.services.length}
              subheader='Total Services'
            />
          </Card>
        </Grid>
        <Grid item xs>
          <Card>
            <CardHeader
              title={noIntegrationTotal}
              subheader='Services Missing Integration'
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
        <Grid item xs>
          <Card>
            <CardHeader
              title={noNotifTotal}
              subheader='Services Missing Notifications'
              action={
                !!noNotifTotal && (
                  <Tooltip title='Services with empty escalation policies.'>
                    <WarningAmberOutlined color='warning' />
                  </Tooltip>
                )
              }
            />
          </Card>
        </Grid>
        <Grid item xs>
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
      </React.Fragment>
    )
  }

  function renderUsageGraphs(): JSX.Element {
    return (
      <React.Fragment>
        <Grid item xs>
          <Card className={styles.card}>
            <CardHeader
              title='Integration Key Usage'
              subheader={
                metrics.filteredServices?.length +
                ' services, of which ' +
                filteredNoIntKeyTotal +
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
          <Card className={styles.card}>
            <CardHeader
              title='Escalation Policy Usage'
              subheader={
                metrics.filteredServices?.length +
                ' services, of which ' +
                filteredNoNotifTotal +
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
      <Card className={styles.card}>
        <CardHeader title='Services' />
        <CardContent>
          <AdminServiceTable
            services={metrics.filteredServices}
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
