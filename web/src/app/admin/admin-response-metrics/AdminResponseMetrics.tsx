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
  formatDuration as dur,
  formatPercent as pct,
  responseByService,
  summarizeResponse,
  ServiceResponse,
} from './util'

const RANGES = [
  { value: 'P7D', label: 'Past 7 days' },
  { value: 'P30D', label: 'Past 30 days' },
  { value: 'P90D', label: 'Past 90 days' },
]

const COLUMNS: [string, (s: ServiceResponse) => string | number][] = [
  ['Alerts', (s) => s.total],
  ['Closed', (s) => s.closed],
  ['Escalated', (s) => pct(s.escalated, s.closed)],
  ['Median time to ack', (s) => dur(s.medianAckMs)],
  ['Median time to close', (s) => dur(s.medianCloseMs)],
]

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
  const { alerts, loading, error } = useAlerts(
    { notCreatedBefore: since },
    since,
  )

  const s = useMemo(() => summarizeResponse(alerts), [alerts])
  const ackBuckets = useMemo(() => ackDistribution(alerts), [alerts])
  const services = useMemo(() => responseByService(alerts), [alerts])

  if (error) return <GenericError error={error.message} />

  const cards = [
    ['Alerts', String(s.total), `${s.closed} closed`],
    ['Open', String(s.open), `${s.unacknowledged} unacknowledged`],
    ['Time to ack', dur(s.medianAckMs), `median · p90 ${dur(s.p90AckMs)}`],
    [
      'Time to close',
      dur(s.medianCloseMs),
      `median · p90 ${dur(s.p90CloseMs)}`,
    ],
    ['Escalated', pct(s.escalated, s.closed), 'of closed alerts'],
    ['Services', String(services.length), 'with alerts'],
  ]

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

      {cards.map(([title, value, subheader]) => (
        <Grid key={title} item xs={6} sm={4} md={2} data-cy='response-stat'>
          <Card sx={{ height: '100%' }}>
            <CardContent>
              <Typography variant='body2' color='text.secondary'>
                {title}
              </Typography>
              <Typography variant='h4' component='div' sx={{ my: 0.5 }}>
                {loading ? '…' : value}
              </Typography>
              <Typography variant='caption' color='text.secondary'>
                {subheader}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      ))}

      <Grid item xs={12} md={5}>
        <Card>
          <CardHeader
            title='Time to Acknowledge'
            subheader='Closed alerts, by how long they took to be acknowledged'
          />
          <CardContent>
            <ResponseBarGraph data={ackBuckets} loading={loading} />
          </CardContent>
        </Card>
      </Grid>

      <Grid item xs={12} md={7}>
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
                  {COLUMNS.map(([c]) => (
                    <TableCell key={c} align='right'>
                      {c}
                    </TableCell>
                  ))}
                </TableRow>
              </TableHead>
              <TableBody>
                {services.map((svc) => (
                  <TableRow key={svc.id} hover>
                    <TableCell>
                      <AppLink to={`/services/${svc.id}/alert-metrics`}>
                        {svc.name}
                      </AppLink>
                    </TableCell>
                    {COLUMNS.map(([c, get]) => (
                      <TableCell key={c} align='right'>
                        {get(svc)}
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
                {!loading && !services.length && (
                  <TableRow>
                    <TableCell colSpan={6} align='center'>
                      No alerts in this time range
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
