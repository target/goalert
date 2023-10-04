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
    noHBTotal: number
    noIKTotal: number
    noEPTotal: number
    alertLimitTotal: number
  } => {
    let noHBTotal = 0
    let noIKTotal = 0
    let noEPTotal = 0
    let alertLimitTotal = 0
    services.map((svc: Service) => {
      if (!svc.heartbeatMonitors.length) noHBTotal += 1
      if (!svc.integrationKeys.length) noIKTotal += 1
      if (!svc.escalationPolicy?.steps.length) noEPTotal += 1
      else {
        let emptyTargets = 0
        svc.escalationPolicy.steps.map((step) => {
          if (!step.targets.length) emptyTargets += 1
          if (emptyTargets === svc.escalationPolicy?.steps.length)
            noEPTotal += 1
        })
      }
      if (svc.notices.length) alertLimitTotal += 1
    })
    return { noHBTotal, noIKTotal, noEPTotal, alertLimitTotal }
  }

  const { noHBTotal, noIKTotal, noEPTotal, alertLimitTotal } =
    getConfigIssueCounts(serviceData.services || [])

  const { noIKTotal: filteredNoIKTotal, noEPTotal: filteredNoEPTotal } =
    getConfigIssueCounts(metrics.filteredServices || [])

  return (
    <Grid container spacing={2}>
      {serviceData.loading && <Spinner />}
      <Grid item xs={3}>
        <Card>
          <CardHeader
            title={serviceData.services.length}
            subheader='Total Services'
          />
        </Card>
      </Grid>
      <Grid item xs={3}>
        <Card>
          <CardHeader
            title={noHBTotal + noIKTotal}
            subheader='Services Missing Integrations'
            action={
              !!(noHBTotal + noIKTotal) && (
                <Tooltip title='Services with no integration keys or heartbeat monitors.'>
                  <WarningAmberOutlined color='warning' />
                </Tooltip>
              )
            }
          />
        </Card>
      </Grid>
      <Grid item xs={3}>
        <Card>
          <CardHeader
            title={noEPTotal}
            subheader='Services Missing Notifications'
            action={
              !!noEPTotal && (
                <Tooltip title='Services with empty escalation policies.'>
                  <WarningAmberOutlined color='warning' />
                </Tooltip>
              )
            }
          />
        </Card>
      </Grid>
      <Grid item xs={3}>
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
        <AdminServiceFilter />
      </Grid>
      <Grid item xs={6}>
        <Card className={styles.card}>
          <CardHeader
            title='Integration Key Usage'
            subheader={
              metrics.filteredServices?.length +
              ' services, of which ' +
              filteredNoIKTotal +
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
      <Grid item xs={6}>
        <Card className={styles.card}>
          <CardHeader
            title='Escalation Policy Usage'
            subheader={
              metrics.filteredServices?.length +
              ' services, of which ' +
              filteredNoEPTotal +
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
      <Grid item xs={12}>
        <Card className={styles.card}>
          <CardHeader title='Services' />
          <CardContent>
            <AdminServiceTable
              services={metrics.filteredServices}
              loading={serviceData.loading}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
