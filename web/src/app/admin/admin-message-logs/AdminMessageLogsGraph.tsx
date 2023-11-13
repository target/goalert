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
import { DateTime, Duration } from 'luxon'
import Spinner from '../../loading/components/Spinner'
import { gql, useQuery } from 'urql'
import { Time } from '../../util/Time'
import { getValidIntervals, useMessageLogsParams } from './util'

type Stats = Array<{
  start: string
  end: string
  count: number
}>

interface MessageLogGraphData {
  date: string
  label: React.ReactNode
  count: number
}

const statsQuery = gql`
  query messageStatsQuery(
    $logsInput: MessageLogSearchOptions
    $statsInput: TimeSeriesOptions!
  ) {
    messageLogs(input: $logsInput) {
      stats {
        timeSeries(input: $statsInput) {
          start
          end
          count
        }
      }
    }
  }
`

export default function AdminMessageLogsGraph(): React.ReactNode {
  const theme = useTheme()

  const [{ search, start, end, graphInterval }, setParams] =
    useMessageLogsParams()

  const [{ data, fetching, error }] = useQuery({
    query: statsQuery,
    variables: {
      logsInput: {
        search,
        createdAfter: start,
        createdBefore: end,
      },
      statsInput: {
        bucketDuration: graphInterval,
        bucketOrigin: start,
      },
    },
  })
  const stats: Stats = data?.messageLogs?.stats?.timeSeries ?? []

  const graphData = React.useMemo(
    (): MessageLogGraphData[] =>
      stats.map(({ start, end, count }) => ({
        count,
        date: start,
        label: (
          <React.Fragment>
            <Time time={start} /> - <Time time={end} />
          </React.Fragment>
        ),
      })),
    [stats],
  )

  const formatIntervals = (label: string): string => {
    if (label.toString() === '0' || label === 'auto') return ''
    const dt = DateTime.fromISO(label)
    const dur = Duration.fromISO(graphInterval)

    if (dur.as('hours') < 1)
      return dt.toLocaleString({
        hour: 'numeric',
        minute: 'numeric',
      })

    if (dur.as('days') < 1) return dt.toLocaleString({ hour: 'numeric' })

    return dt.toLocaleString({
      month: 'short',
      day: 'numeric',
    })
  }

  return (
    <Grid item xs={12}>
      <Accordion defaultExpanded>
        <AccordionSummary expandIcon={<ExpandMoreIcon />}>
          <CardHeader
            title='Message Logs'
            subheader={
              (fetching && <Spinner text='Loading...' />) ||
              (error && (
                <Typography>Error loading graph: {error.message}</Typography>
              ))
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
                <InputLabel id='interval-select-label' sx={{ pr: 1 }}>
                  Interval Duration
                </InputLabel>
                <Select
                  labelId='interval-select-label'
                  id='interval-select'
                  value={graphInterval}
                  onChange={(e) => setParams({ graphInterval: e.target.value })}
                >
                  {getValidIntervals({ start, end }).map((ivl) => (
                    <MenuItem key={ivl.value} value={ivl.value}>
                      {ivl.label}
                    </MenuItem>
                  ))}
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
                {({ width, height }: { width: number; height: number }) => (
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
