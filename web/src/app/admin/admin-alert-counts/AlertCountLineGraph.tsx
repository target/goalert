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

interface AlertCountLineGraphProps {
  data: (typeof LineChart.defaultProps)['data']
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

interface CustomDotProps extends DotProps {
  payload: {
    date: string
    dayTotal: number
  }
}

const CustomDot = (props: CustomDotProps): JSX.Element => {
  const { cy, cx, fill, r, stroke, strokeWidth, name, payload } = props
  return (
    <circle
      cy={cy}
      cx={cx}
      fill={fill}
      r={payload.dayTotal ? r : 0}
      stroke={stroke}
      strokeWidth={strokeWidth}
      key={name + '-' + payload.date}
      data-cy={name + '-' + payload.date}
    />
  )
}

export default function AlertCountLineGraph(
  props: AlertCountLineGraphProps,
): JSX.Element {
  const [active, setActive] = useState('')
  const classes = useStyles()
  const theme = useTheme()

  const chooseColor = (idx: number): string => {
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

  const formatTick = (label: string): string => {
    // check for default bounds
    if (label.toString() !== '0' && label !== 'auto') {
      const dateTime = DateTime.fromFormat(label, 'MMM d, t')
      if (props.unit === 'month') return dateTime.toFormat('MMM')
      if (props.unit === 'week' || props.unit === 'day')
        return dateTime.toFormat('MMM d')
      if (props.unit === 'hour' || props.unit === 'minute')
        return dateTime.toFormat('t')
    }
    return ''
  }

  return (
    <Grid container className={classes.graphContent}>
      <Grid item xs={12} data-cy='alert-count-graph'>
        {props.loading && <Spinner />}
        <AutoSizer>
          {({ width, height }) => (
            <LineChart
              width={width}
              height={height}
              data={props.data}
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
                dataKey='date'
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
                content={({ active, payload, label }) => {
                  if (!active || !payload?.length) return null
                  return (
                    <Paper variant='outlined' sx={{ p: 1 }}>
                      <Typography variant='body2'>{label}</Typography>
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
                  dataKey='dayTotal'
                  data={series.dailyCounts}
                  strokeOpacity={props.loading ? 0.5 : 1}
                  strokeWidth={active === series.serviceName ? 4 : 1}
                  name={series.serviceName}
                  stroke={chooseColor(idx)}
                  dot={CustomDot}
                  key={idx}
                />
              ))}
            </LineChart>
          )}
        </AutoSizer>
      </Grid>
    </Grid>
  )
}
