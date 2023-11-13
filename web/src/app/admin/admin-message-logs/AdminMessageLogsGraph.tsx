import React, { useMemo } from 'react'
import { useTheme } from '@mui/material/styles'
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  CardContent,
  CardHeader,
  FormControl,
  FormControlLabel,
  FormLabel,
  Grid,
  IconButton,
  InputLabel,
  MenuItem,
  Paper,
  Radio,
  RadioGroup,
  Select,
  Typography,
} from '@mui/material'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import ResetIcon from '@mui/icons-material/Replay'
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
import { TimeSeriesBucket } from '../../../schema'

const statsQuery = gql`
  query messageStatsQuery(
    $logsInput: MessageLogSearchOptions
    $statsInput: TimeSeriesOptions!
    $segmentBy: MessageLogSegmentBy
  ) {
    messageLogs(input: $logsInput) {
      stats {
        timeSeries(input: $statsInput, segmentBy: $segmentBy) {
          start
          end
          count
          segmentLabel
        }
      }
    }
  }
`

export default function AdminMessageLogsGraph(): JSX.Element {
  const theme = useTheme()

  const [{ search, start, end, graphInterval, segmentBy }, setParams] =
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
      segmentBy: segmentBy || null,
    },
  })
  const stats: TimeSeriesBucket[] = data?.messageLogs?.stats?.timeSeries ?? []

  // get list of segment labels from data to map out each line on the graph
  // set one line of "Message Counts" if user hasn't selected a segmentBy option
  const graphLines = useMemo(() => {
    const gl: { [key: string]: TimeSeriesBucket[] } = {}
    if (!segmentBy) return { 'Message Counts': stats }

    stats.forEach((stat) => {
      if (!gl[stat.segmentLabel]) gl[stat.segmentLabel] = [stat]
      else gl[stat.segmentLabel].push(stat)
    })

    return gl
  }, [stats])

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

  function handleOnSegmentByChange(
    event: React.ChangeEvent<HTMLInputElement>,
  ): void {
    setParams({
      segmentBy: (event.target as HTMLInputElement).value,
    })
  }

  function getLineStroke(index: number): string {
    switch (index) {
      case 0:
        return theme.palette.primary.main
      case 1:
        return theme.palette.secondary.main
      default:
        return `hsl(${index * 137.508},60%,75%)`
    }
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
              )) || (
                <Typography color='textSecondary'>
                  Click on a point to view more information
                </Typography>
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
                sx={{ display: 'grid', alignItems: 'center', pt: 0, pr: 3 }}
              >
                <InputLabel id='interval-select-label' sx={{ pb: 1 }}>
                  Interval Duration
                </InputLabel>
                <Select
                  labelId='interval-select-label'
                  id='interval-select'
                  value={graphInterval}
                  onChange={(e) => setParams({ graphInterval: e.target.value })}
                  size='small'
                >
                  {getValidIntervals({ start, end }).map((ivl) => (
                    <MenuItem key={ivl.value} value={ivl.value}>
                      {ivl.label}
                    </MenuItem>
                  ))}
                </Select>
              </CardContent>
            </Grid>
            <Grid item>
              <FormControl>
                <FormLabel id='segment-by' sx={{ pb: 1 }}>
                  Segment By
                </FormLabel>
                <RadioGroup
                  row
                  aria-labelledby='segment-by'
                  name='segment-by-group'
                  value={segmentBy}
                  onChange={handleOnSegmentByChange}
                >
                  <FormControlLabel
                    value='service'
                    control={<Radio />}
                    label='Service'
                  />
                  <FormControlLabel
                    value='user'
                    control={<Radio />}
                    label='User'
                  />
                  <FormControlLabel
                    value='messageType'
                    control={<Radio />}
                    label='Message Type'
                  />
                </RadioGroup>
              </FormControl>
            </Grid>
            <Grid item sx={{ display: 'flex', alignItems: 'center' }}>
              <IconButton
                onClick={() =>
                  setParams({
                    segmentBy: '',
                  })
                }
              >
                <ResetIcon />
              </IconButton>
            </Grid>
            <Grid
              item
              xs={12}
              data-cy='message-logs-graph'
              sx={{ height: 500 }}
            >
              <AutoSizer>
                {({ width, height }: { width: number; height: number }) =>
                  stats.length === 0 && !fetching ? (
                    <div
                      style={{
                        width,
                        height,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                      }}
                    >
                      No Results
                    </div>
                  ) : (
                    <LineChart
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
                        dataKey='start'
                        type='category'
                        stroke={theme.palette.text.secondary}
                        tickFormatter={formatIntervals}
                        minTickGap={15}
                        allowDuplicatedCategory={false}
                      />
                      <YAxis
                        type='number'
                        dataKey='count'
                        allowDecimals={false}
                        interval='preserveStart'
                        stroke={theme.palette.text.secondary}
                      />
                      <Tooltip
                        trigger='click'
                        cursor={{ fill: theme.palette.background.default }}
                        content={(data) => {
                          const { active, payload } = data
                          if (!active || !payload?.length) return null

                          return (
                            <Paper
                              data-cy='message-log-tooltip'
                              variant='outlined'
                              sx={{
                                p: 1,
                                maxHeight: '250px',
                                overflow: 'auto',
                                pointerEvents: 'auto',
                              }}
                            >
                              <Typography variant='body2' sx={{ pb: 1 }}>
                                <Time time={payload[0].payload.start} /> -{' '}
                                <Time time={payload[0].payload.end} />
                              </Typography>
                              {payload.map((p) => {
                                if (p.payload.count === 0) return null

                                return (
                                  <React.Fragment key={p.name}>
                                    <Typography variant='body2'>
                                      {p.payload.segmentLabel === ''
                                        ? 'Count'
                                        : p.payload.segmentLabel}
                                      : {p.payload.count}
                                    </Typography>
                                  </React.Fragment>
                                )
                              })}
                            </Paper>
                          )
                        }}
                      />
                      <Legend
                        wrapperStyle={{
                          maxHeight: '100px',
                          overflow: 'auto',
                        }}
                      />
                      {Object.keys(graphLines).map((label, index) => (
                        <Line
                          key={label}
                          dataKey='count'
                          data={graphLines[label]}
                          name={label}
                          type='monotone'
                          stroke={getLineStroke(index)}
                          strokeWidth={2}
                          isAnimationActive={false}
                          dot={(props) => (
                            <circle {..._.omit(props, 'dataKey')} />
                          )}
                        />
                      ))}
                    </LineChart>
                  )
                }
              </AutoSizer>
            </Grid>
          </Grid>
        </AccordionDetails>
      </Accordion>
    </Grid>
  )
}
