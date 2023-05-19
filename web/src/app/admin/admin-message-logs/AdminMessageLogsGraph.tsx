import React from 'react'
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
import { DateTime, DateTimeFormatOptions } from 'luxon'
import { useURLParam } from '../../actions'
import Spinner from '../../loading/components/Spinner'
import { CombinedError } from 'urql'

interface AdminMessageLogsGraphProps {
  loading: boolean
  error?: CombinedError
  stats: Array<{
    start: string
    end: string
    count: number
  }>
}

interface MessageLogGraphData {
  date: string
  label: string
  count: number
}

export default function AdminMessageLogsGraph({
  loading,
  // error,
  stats = [],
}: AdminMessageLogsGraphProps): JSX.Element {
  const theme = useTheme()

  // graph duration set with ISO duration values, e.g. PT8H for a duration of 8 hours
  const [duration, setDuration] = useURLParam<string>('graphInterval', 'PT1H')

  const locale: DateTimeFormatOptions = {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
  }
  const graphData = React.useMemo(
    (): MessageLogGraphData[] =>
      stats.map(({ start, end, count }) => ({
        count,
        date: start,
        label:
          DateTime.fromISO(start).toLocaleString(locale) +
          ' - ' +
          DateTime.fromISO(end).toLocaleString(locale),
      })),
    [stats],
  )

  const formatIntervals = (label: string): string => {
    // check for default bounds
    if (label.toString() !== '0' && label !== 'auto') {
      const dt = DateTime.fromFormat(label, 'MMM d, t')
      if (duration.endsWith('D')) return dt.toFormat('MMM d')
      if (duration.endsWith('H')) return dt.toFormat('h a')
      if (duration.endsWith('M')) return dt.toFormat('h:mma').toLowerCase()
    }
    return ''
  }

  return (
    <Grid item xs={12}>
      <Accordion defaultExpanded>
        <AccordionSummary expandIcon={<ExpandMoreIcon />}>
          <CardHeader
            title='Message Logs'
            subheader={loading && <Spinner text='Loading...' />}
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
                <InputLabel id='interval-select-label' sx={{ pr: 1 }}>
                  Interval Duration
                </InputLabel>
                <Select
                  labelId='interval-select-label'
                  id='interval-select'
                  value={duration}
                  onChange={(e) => setDuration(e.target.value)}
                >
                  <MenuItem value='P1D'>Daily</MenuItem>
                  <MenuItem value='PT8H'>8 hours</MenuItem>
                  <MenuItem value='PT1H'>Hourly</MenuItem>
                  <MenuItem value='PT15M'>15 minutes</MenuItem>
                  <MenuItem value='PT5M'>5 minutes</MenuItem>
                </Select>
              </CardContent>
            </Grid>
            <Grid
              item
              xs={12}
              data-cy='message-logs-graph'
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
                      cursor={{ fill: theme.palette.background.default }}
                      content={({ active, payload }) => {
                        if (!active || !payload?.length) return null

                        const p = payload[0].payload
                        return (
                          <Paper
                            data-cy='message-log-tooltip'
                            variant='outlined'
                            sx={{ p: 1 }}
                          >
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
