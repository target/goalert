import React, { useState } from 'react'
import { Grid, Paper, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles/makeStyles'
import { Theme, useTheme } from '@mui/material/styles'
import { DateTime, DateTimeUnit } from 'luxon'
import {
  blueGrey,
  teal,
  green,
  cyan,
  amber,
  pink,
  brown,
} from '@mui/material/colors'
import AutoSizer from 'react-virtualized-auto-sizer'
import {
  XAxis,
  YAxis,
  CartesianGrid,
  LineChart,
  Line,
  DotProps,
  Legend,
  Tooltip,
} from 'recharts'
import Spinner from '../../loading/components/Spinner'
import { Time } from '../../util/Time'

type DataType = React.ComponentProps<typeof LineChart>['data']

interface AlertCountLineGraphProps {
  data: DataType
  loading: boolean
  unit: DateTimeUnit
}

const useStyles = makeStyles((theme: Theme) => ({
  graphContent: {
    height: '500px',
    fontFamily: theme.typography.body2.fontFamily,
  },
  bar: {
    '&:hover': {
      cursor: 'pointer',
    },
  },
}))

// Symbol used to store the timestamp in the payload without
// interfering with the data keys.
const TIMESTAMP_SYM = Symbol('timestamp')
const getTimestamp = (p: CustomDotProps['payload']): string => p[TIMESTAMP_SYM]

interface CustomDotProps extends DotProps {
  payload: {
    [TIMESTAMP_SYM]: string
    [key: string]: number | string
  }
}

function CustomDot(props: CustomDotProps): JSX.Element {
  const { cy, cx, fill, r, stroke, strokeWidth, name = '', payload } = props

  return (
    <circle
      cy={cy}
      cx={cx}
      fill={fill}
      r={payload[name] === 0 ? 0 : r}
      stroke={stroke}
      strokeWidth={strokeWidth}
      key={name + '-' + payload[TIMESTAMP_SYM]}
      data-cy={name + '-' + payload[TIMESTAMP_SYM]}
    />
  )
}

export default function AlertCountLineGraph(
  props: AlertCountLineGraphProps,
): JSX.Element {
  const [active, setActive] = useState('')
  const classes = useStyles()
  const theme = useTheme()

  function chooseColor(idx: number): string {
    const shade = theme.palette.mode === 'light' ? 'A700' : 'A400'

    switch (idx) {
      case 1:
        return teal[shade]
      case 2:
        return brown[shade]
      case 3:
        return green[shade]
      case 4:
        return cyan[shade]
      case 5:
        return amber[shade]
      case 6:
        return pink[shade]
      default:
        return blueGrey[shade]
    }
  }

  function formatTick(date: string): string {
    const dt = DateTime.fromISO(date)
    // check for default bounds

    if (props.unit === 'month') return dt.toLocaleString({ month: 'long' })
    if (props.unit === 'week' || props.unit === 'day')
      return dt.toLocaleString({ month: 'short', day: 'numeric' })
    if (props.unit === 'hour') return dt.toLocaleString({ hour: 'numeric' })
    if (props.unit === 'minute')
      return dt.toLocaleString({ hour: 'numeric', minute: 'numeric' })

    return date
  }

  function flattenData(
    data: DataType,
  ): Array<{ [key: string]: number | string }> {
    const dateMap: { [key: string]: { [key: string]: number | string } } = {}

    if (!data) return []

    // Populate the map with data and flatten the structure in a single pass
    data.forEach((service) => {
      service.dailyCounts.forEach(
        (dailyCount: { date: string; dayTotal: number }) => {
          if (!dateMap[dailyCount.date]) {
            dateMap[dailyCount.date] = { [TIMESTAMP_SYM]: dailyCount.date }
          }
          dateMap[dailyCount.date][service.serviceName] = dailyCount.dayTotal
        },
      )
    })

    // Convert the map into the desired array structure
    return Object.values(dateMap)
  }

  const data = React.useMemo(() => flattenData(props.data), [props.data])

  return (
    <Grid container className={classes.graphContent}>
      <Grid item xs={12} data-cy='alert-count-graph'>
        {props.loading && <Spinner />}
        <AutoSizer>
          {({ width, height }: { width: number; height: number }) => (
            <LineChart
              width={width}
              height={height}
              data={data}
              margin={{
                top: 50,
                right: 30,
                bottom: 50,
              }}
            >
              <CartesianGrid
                strokeDasharray='4'
                vertical={false}
                stroke={theme.palette.text.secondary}
              />
              <XAxis
                dataKey={getTimestamp}
                allowDuplicatedCategory={false}
                minTickGap={15}
                stroke={theme.palette.text.secondary}
                interval='preserveStartEnd'
                tickFormatter={formatTick}
              />
              <YAxis
                allowDecimals={false}
                interval='preserveStart'
                stroke={theme.palette.text.secondary}
              />
              <Tooltip
                data-cy='alert-count-tooltip'
                cursor={{ fill: theme.palette.background.default }}
                content={({ active, payload = [], label }) => {
                  if (!active || !payload?.length) return null
                  return (
                    <Paper variant='outlined' sx={{ p: 1 }}>
                      <Typography variant='body2'>
                        <Time time={label} />
                      </Typography>
                      {payload.map((svc, idx) => {
                        return (
                          <React.Fragment key={idx}>
                            <Typography
                              variant='body2'
                              color={chooseColor(idx)}
                            >{`${svc.name}: ${svc.value}`}</Typography>
                          </React.Fragment>
                        )
                      })}
                    </Paper>
                  )
                }}
              />
              <Legend
                onMouseEnter={(e) => {
                  setActive(e.value)
                }}
                onMouseLeave={() => {
                  setActive('')
                }}
              />
              {props.data?.map((series, idx) => (
                <Line
                  key={series.serviceName + idx}
                  dataKey={series.serviceName}
                  strokeOpacity={props.loading ? 0.5 : 1}
                  strokeWidth={active === series.serviceName ? 4 : 1}
                  name={series.serviceName}
                  stroke={chooseColor(idx)}
                  dot={CustomDot}
                />
              ))}
            </LineChart>
          )}
        </AutoSizer>
      </Grid>
    </Grid>
  )
}
