import React, { useMemo } from 'react'
import {
  Card,
  CardContent,
  CardHeader,
  Grid,
  MenuItem,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material'
import { DateTime, Duration } from 'luxon'
import { useURLParams } from '../../actions'
import { useAlerts } from '../../services/AlertMetrics/useAlerts'
import { GenericError } from '../../error-pages'
import AppLink from '../../util/AppLink'
import ResponseBarGraph from './ResponseBarGraph'
import {
  ackDistribution,
  alertsByHour,
  formatDuration,
  formatPercent,
  responseByService,
  summarizeResponse,
} from './util'

const RANGES = [
  { value: 'P7D', label: 'Past 7 days' },
  { value: 'P30D', label: 'Past 30 days' },
  { value: 'P90D', label: 'Past 90 days' },
]

interface StatCardProps {
  title: string
  value: string
  subheader: string
}

function StatCard(props: StatCardProps): React.JSX.Element {
  return (
    <Card sx={{ height: '100%' }}>
      <CardContent>
        <Typography variant='body2' color='text.secondary'>
          {props.title}
        </Typography>
        <Typography variant='h4' component='div' sx={{ my: 0.5 }}>
          {props.value}
        </Typography>
        <Typography variant='caption' color='text.secondary'>
          {props.subheader}
        </Typography>
      </CardContent>
    </Card>
  )
}

export default function AdminResponseMetrics(): React.JSX.Element {
  const [params, setParams] = useURLParams({ range: 'P30D' })
  const range = RANGES.find((r) => r.value === params.range) ?? RANGES[1]

  const since = useMemo(
    () =>
      DateTime.now()
        .minus(Duration.fromISO(range.value))
        .startOf('minute')
        .toISO(),
    [range.value],
  )
  const alertsData = useAlerts({ notCreatedBefore: since }, since)
  const { alerts, loading } = alertsData

  const summary = useMemo(() => summarizeResponse(alerts), [alerts])
  const ackBuckets = useMemo(() => ackDistribution(alerts), [alerts])
  const hours = useMemo(() => alertsByHour(alerts), [alerts])
  const services = useMemo(() => responseByService(alerts), [alerts])

  if (alertsData.error) {
    return <GenericError error={alertsData.error.message} />
  }

  const show = (v: string): string => (loading ? '…' : v)

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <TextField
          select
          size='small'
          label='Time range'
          value={range.value}
          onChange={(e) => setParams({ range: e.target.value })}
          sx={{ minWidth: 200 }}
          data-cy='response-range'
        >
          {RANGES.map((r) => (
            <MenuItem key={r.value} value={r.value}>
              {r.label}
            </MenuItem>
          ))}
        </TextField>
      </Grid>

      <Grid item xs={6} sm={4} md={2} data-cy='response-total'>
        <StatCard
          title='Alerts'
          value={show(String(summary.total))}
          subheader={`${summary.closed} closed`}
        />
      </Grid>
      <Grid item xs={6} sm={4} md={2}>
        <StatCard
          title='Open'
          value={show(String(summary.open))}
          subheader={`${summary.unacknowledged} unacknowledged`}
        />
      </Grid>
      <Grid item xs={6} sm={4} md={2}>
        <StatCard
          title='Time to ack'
          value={show(formatDuration(summary.medianAckMs))}
          subheader={`median · p90 ${formatDuration(summary.p90AckMs)}`}
        />
      </Grid>
      <Grid item xs={6} sm={4} md={2}>
        <StatCard
          title='Time to close'
          value={show(formatDuration(summary.medianCloseMs))}
          subheader={`median · p90 ${formatDuration(summary.p90CloseMs)}`}
        />
      </Grid>
      <Grid item xs={6} sm={4} md={2}>
        <StatCard
          title='Escalated'
          value={show(formatPercent(summary.escalated, summary.closed))}
          subheader='of closed alerts'
        />
      </Grid>
      <Grid item xs={6} sm={4} md={2}>
        <StatCard
          title='Services'
          value={show(String(services.length))}
          subheader='with alerts'
        />
      </Grid>

      <Grid item xs={12} md={6}>
        <Card>
          <CardHeader
            title='Time to Acknowledge'
            subheader='Closed alerts, by how long they took to be acknowledged'
          />
          <CardContent>
            <ResponseBarGraph
              data={ackBuckets}
              loading={loading}
              showAllLabels
              data-cy='response-ack-graph'
            />
          </CardContent>
        </Card>
      </Grid>
      <Grid item xs={12} md={6}>
        <Card>
          <CardHeader
            title='Alerts by Hour of Day'
            subheader='When alerts are created, in your local time'
          />
          <CardContent>
            <ResponseBarGraph
              data={hours}
              loading={loading}
              data-cy='response-hour-graph'
            />
          </CardContent>
        </Card>
      </Grid>

      <Grid item xs={12}>
        <Card>
          <CardHeader
            title='Response by Service'
            subheader={`${range.label}. Time metrics include closed alerts only.`}
          />
          <CardContent sx={{ overflowX: 'auto' }}>
            <Table size='small' data-cy='response-service-table'>
              <TableHead>
                <TableRow>
                  <TableCell>Service</TableCell>
                  <TableCell align='right'>Alerts</TableCell>
                  <TableCell align='right'>Closed</TableCell>
                  <TableCell align='right'>Escalated</TableCell>
                  <TableCell align='right'>Median time to ack</TableCell>
                  <TableCell align='right'>Median time to close</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {services.map((s) => (
                  <TableRow key={s.id} hover>
                    <TableCell>
                      <AppLink to={`/services/${s.id}/alert-metrics`}>
                        {s.name}
                      </AppLink>
                    </TableCell>
                    <TableCell align='right'>{s.total}</TableCell>
                    <TableCell align='right'>{s.closed}</TableCell>
                    <TableCell align='right'>
                      {formatPercent(s.escalated, s.closed)}
                    </TableCell>
                    <TableCell align='right'>
                      {formatDuration(s.medianAckMs)}
                    </TableCell>
                    <TableCell align='right'>
                      {formatDuration(s.medianCloseMs)}
                    </TableCell>
                  </TableRow>
                ))}
                {!loading && !services.length && (
                  <TableRow>
                    <TableCell colSpan={6} align='center'>
                      <Typography variant='body2' color='text.secondary'>
                        No alerts in this time range
                      </Typography>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
