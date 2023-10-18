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
  VolumeOffOutlined,
} from '@mui/icons-material'
import { Service } from '../../../schema'

const useStyles = makeStyles((theme: Theme) => ({
  card: {
    marginTop: theme.spacing(1),
  },
}))

export default function AdminServiceMetrics(): JSX.Element {
  const now = useMemo(() => DateTime.now(), [])
  const styles = useStyles()

  const depKey = `${now}`
  const serviceData = useServices(depKey)
  const [metrics] = useWorker(
    'useServiceMetrics',
    {
      services: serviceData.services,
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
            title={noIntegrationTotal}
            subheader='Services Missing Integrations'
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
      <Grid item xs={3}>
        <Card>
          <CardHeader
            title={noEPTotal}
            subheader='Services Missing Notifications'
            action={
              !!noEPTotal && (
                <Tooltip title='Services with empty escalation policies.'>
                  <VolumeOffOutlined color='warning' />
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
