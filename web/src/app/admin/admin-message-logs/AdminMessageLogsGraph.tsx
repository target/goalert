import React, { useMemo } from 'react'
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
import { DateTime } from 'luxon'
import { DebugMessage } from '../../../schema'
import { useWorker } from '../../worker'
import Spinner from '../../loading/components/Spinner'

export default function AdminMessageLogsGraph(props: {
  logs: DebugMessage[]
  loadingData: boolean
}): JSX.Element {
  const { logs, loadingData } = props
  const theme = useTheme()
  const [params] = useURLParams({
    search: '',
    start: '',
    end: '',
  })

  // graph duration set with ISO duration values, e.g. PT8H for a duration of 8 hours
  const [duration, setDuration] = useURLParam<string>('graphInterval', 'PT1H')

  const opts = useMemo(
    () => ({ start: params.start, end: params.end, duration, logs }),
    [params.start, params.end, duration, logs],
  )
  const [graphData, status] = useWorker('useMessageLogGraphData', opts, [])
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
            subheader={
              (status.loading || loadingData) && (
                <Spinner text={`Total Loaded: ${logs.length}`} />
              )
            }
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
