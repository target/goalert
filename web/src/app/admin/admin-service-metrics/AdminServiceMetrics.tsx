import React, { useMemo } from 'react'
import { Grid, Card, CardHeader, CardContent } from '@mui/material'
import { DateTime } from 'luxon'
import { useServices } from './useServices'
import { ServiceSearchOptions } from '../../../schema'
import { useWorker } from '../../worker'
import { ServiceMetrics } from './useServiceMetrics'
import AdminServiceKeyGraph from './AdminServiceKeyGraph'
import AdminServiceEPGraph from './AdminServiceEPGraph'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import AdminServiceKeyTable from './AdminServiceKeyTable'
import AdminServiceEPTable from './AdminServiceEPTable'

const useStyles = makeStyles((theme: Theme) => ({
  card: {
    marginTop: theme.spacing(1),
  },
}))

export default function AdminServiceMetrics(): JSX.Element {
  const now = useMemo(() => DateTime.now(), [])
  const styles = useStyles()

  const depKey = `${now}`
  const serviceData = useServices({} as ServiceSearchOptions, depKey)
  const [metrics] = useWorker(
    'useServiceMetrics',
    { services: serviceData.services },
    {} as ServiceMetrics,
  )
  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Card className={styles.card}>
          <CardHeader
            title='Integration Key Metrics'
            subheader={
              serviceData.services?.length +
              ' total services, of which ' +
              metrics.noIntKeys?.length +
              ' service(s) do not have any integration keys configured.'
            }
          />
          <CardContent>
            <AdminServiceKeyGraph metrics={metrics.intKeyTargets} />
            <AdminServiceKeyTable
              services={serviceData.services}
              loading={serviceData.loading}
            />
          </CardContent>
        </Card>
      </Grid>
      <Grid item xs={12}>
        <Card className={styles.card}>
          <CardHeader
            title='Escalation Policy Metrics'
            subheader={
              serviceData.services?.length +
              ' total services, of which ' +
              metrics.noEPSteps?.length +
              ' service(s) have empty escalation policies.'
            }
          />
          <CardContent>
            <AdminServiceEPGraph metrics={metrics.epStepTargets} />
            <AdminServiceEPTable
              services={serviceData.services}
              loading={serviceData.loading}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
