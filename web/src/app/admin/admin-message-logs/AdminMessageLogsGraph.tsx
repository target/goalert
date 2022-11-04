import React from 'react'
import { useQuery } from '@apollo/client'
import { useTheme } from '@mui/material/styles'
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  CardContent,
  CardHeader,
  Grid,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Typography,
} from '@mui/material'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  LineChart,
  Line,
  Legend,
} from 'recharts'
import AutoSizer from 'react-virtualized-auto-sizer'
import _ from 'lodash'
import { useURLParam, useURLParams } from '../../actions'
import {
  DateTime,
  DateTimeFormatOptions,
  Duration,
  DurationLike,
  Interval,
} from 'luxon'
import { query } from './AdminMessageLogsLayout'
import { DebugMessage } from '../../../schema'

export default function AdminMessageLogsGraph(): JSX.Element {
  const theme = useTheme()
  const [params] = useURLParams({
    search: '',
    start: '',
    end: '',
  })
  const { data, loading, error } = useQuery(query, {
    fetchPolicy: 'cache-first',
    variables: {
      input: {
        search: params.search,
        createdAfter: params.start || null,
        createdBefore: params.end || null,
      },
    },
  })
  // graph duration set with ISO duration values, e.g. P1D for a daily duration
  const [duration, setDuration] = useURLParam<string>('graphInterval', 'P1D')

  if (loading) return <React.Fragment />
  if (error) {
    console.error(error.message)
    return <React.Fragment />
  }
  const logs: DebugMessage[] = data?.data?.nodes ?? []

  // adds a bit of time before/after the first and last alerts
  // if start or end isn't set
  let bufferLen: DurationLike = {}
  if (duration.endsWith('D')) {
    bufferLen = { days: 1 }
  }
  if (duration.endsWith('W')) {
    bufferLen = { weeks: 1 }
  }
  if (duration.endsWith('H')) {
    bufferLen = { hours: 1 }
  }
  if (duration.endsWith('M')) {
    bufferLen = { minutes: 15 }
  }
  const ttlInterval = Interval.fromDateTimes(
    DateTime.fromISO(params.start || logs[logs.length - 1].createdAt).minus(
      bufferLen,
    ),
    DateTime.fromISO(params.end || logs[0].createdAt).plus(bufferLen),
  )

  const intervals = ttlInterval?.splitBy(Duration.fromISO(duration)) ?? []

  const graphData = intervals.map((interval) => {
    const locale: DateTimeFormatOptions = {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: 'numeric',
    }
    const date = interval.start.toLocaleString({
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: 'numeric',
    })
    const label =
      interval.start.toLocaleString(locale) +
      ' - ' +
      interval.end.toLocaleString(locale)

    const intervalLogs = logs.filter((log: DebugMessage) =>
      interval.contains(DateTime.fromISO(log.createdAt)),
    )

    return {
      date,
      label,
      count: intervalLogs.length,
    }
  })

  const formatIntervals = (label: string): string => {
    // check for default bounds
    if (label.toString() !== '0' && label !== 'auto') {
      const dt = DateTime.fromFormat(label, 'MMM d, t')
      if (duration === 'P1D') return dt.toFormat('MMM d')
      if (duration === 'PT1H') return dt.toFormat('h a')
      if (duration === 'PT5M' || duration === 'PT15M')
        return dt.toFormat('h:mma').toLowerCase()
    }
    return ''
  }

  return (
    <Grid item xs={12}>
      <Accordion defaultExpanded>
        <AccordionSummary expandIcon={<ExpandMoreIcon />}>
          <CardHeader
            title='Message Logs'
            subheader={`Total Loaded: ${logs.length}`}
          />
        </AccordionSummary>
        <AccordionDetails>
          <Grid
            container
            sx={{
              fontFamily: theme.typography.body2.fontFamily,
            }}
          >
            <Grid item>
              <CardContent
                sx={{ display: 'flex', alignItems: 'center', pt: 0 }}
              >
                <InputLabel id='demo-simple-select-label' sx={{ pr: 1 }}>
                  Interval Duration
                </InputLabel>
                <Select
                  labelId='demo-simple-select-label'
                  id='demo-simple-select'
                  value={duration}
                  onChange={(e) => setDuration(e.target.value)}
                >
                  <MenuItem value='P1D'>Daily</MenuItem>
                  <MenuItem value='PT1H'>Hourly</MenuItem>
                  <MenuItem value='PT15M'>15 minutes</MenuItem>
                  <MenuItem value='PT5M'>5 minutes</MenuItem>
                </Select>
              </CardContent>
            </Grid>
            <Grid
              item
              xs={12}
              data-cy='metrics-averages-graph'
              sx={{ height: 500 }}
            >
              <AutoSizer>
                {({ width, height }) => (
                  <LineChart
                    data={graphData}
                    width={width}
                    height={height}
                    margin={{
                      top: 30,
                      right: 50,
                    }}
                  >
                    <CartesianGrid
                      strokeDasharray='4'
                      vertical={false}
                      stroke={theme.palette.text.secondary}
                    />
                    <XAxis
                      dataKey='date'
                      type='category'
                      stroke={theme.palette.text.secondary}
                      tickFormatter={formatIntervals}
                      minTickGap={15}
                    />
                    <YAxis
                      type='number'
                      dataKey='count'
                      allowDecimals={false}
                      interval='preserveStart'
                      stroke={theme.palette.text.secondary}
                    />
                    <Tooltip
                      data-cy='message-count-tooltip'
                      cursor={{ fill: theme.palette.background.default }}
                      content={({ active, payload }) => {
                        if (!active || !payload?.length) return null

                        const p = payload[0].payload
                        return (
                          <Paper variant='outlined' sx={{ p: 1 }}>
                            <Typography variant='body2'>{p.label}</Typography>
                            <Typography variant='body2'>
                              Count: {p.count}
                            </Typography>
                          </Paper>
                        )
                      }}
                    />
                    <Legend />
                    <Line
                      type='monotone'
                      strokeWidth={2}
                      stroke={theme.palette.primary.main}
                      activeDot={{ r: 8 }}
                      isAnimationActive={false}
                      dot={(props) => <circle {..._.omit(props, 'dataKey')} />}
                      name='Message Counts'
                      dataKey='count'
                    />
                  </LineChart>
                )}
              </AutoSizer>
            </Grid>
          </Grid>
        </AccordionDetails>
      </Accordion>
    </Grid>
  )
}
