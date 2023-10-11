import React from 'react'
import { Grid, Paper, Typography } from '@mui/material'
import { useTheme } from '@mui/material/styles'
import AutoSizer from 'react-virtualized-auto-sizer'
import Spinner from '../../loading/components/Spinner'
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  LineChart,
  Line,
  Legend,
  DotProps,
} from 'recharts'

interface CustomDotProps extends DotProps {
  dataKey: string
  payload: {
    date: string
    count: number
  }
}

const CustomDot = (props: CustomDotProps): JSX.Element => {
  const { cy, cx, fill, r, stroke, strokeWidth, dataKey, payload } = props
  return (
    <circle
      cy={cy}
      cx={cx}
      fill={fill}
      r={payload.count ? r : 0}
      stroke={stroke}
      strokeWidth={strokeWidth}
      key={dataKey + '-' + payload.date}
      data-cy={dataKey + '-' + payload.date}
    />
  )
}

interface AlertAveragesGraphProps {
  data: (typeof LineChart.defaultProps)['data']
  loading: boolean
}

export default function AlertAveragesGraph(
  props: AlertAveragesGraphProps,
): JSX.Element {
  const theme = useTheme()

  return (
    <Grid
      container
      sx={{
        height: '500px',
        fontFamily: theme.typography.body2.fontFamily,
      }}
    >
      <Grid item xs={12} data-cy='metrics-averages-graph'>
        {props.loading && <Spinner />}
        <AutoSizer>
          {({ width, height }: { width: number; height: number }) => (
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
                type='category'
                stroke={theme.palette.text.secondary}
              />
              <YAxis
                type='number'
                allowDecimals={false}
                interval='preserveStart'
                stroke={theme.palette.text.secondary}
              />
              <Tooltip
                data-cy='metrics-tooltip'
                cursor={{ fill: theme.palette.background.default }}
                content={({ active, payload, label }) => {
                  if (!active || !payload?.length) return null

                  const ackAvg = `${payload[0].name}: ${Math.round(
                    payload[0].payload.avgTimeToAck,
                  )} min`
                  const closeAvg = `${payload[1].name}: ${Math.round(
                    payload[1].payload.avgTimeToClose,
                  )} min`
                  return (
                    <Paper variant='outlined' sx={{ p: 1 }}>
                      <Typography variant='body2'>{label}</Typography>
                      <Typography variant='body2'>{closeAvg}</Typography>
                      <Typography variant='body2'>{ackAvg}</Typography>
                    </Paper>
                  )
                }}
              />
              <Legend />
              <Line
                type='monotone'
                dataKey='avgTimeToAck'
                strokeOpacity={props.loading ? 0.5 : 1}
                strokeWidth={2}
                stroke={theme.palette.primary.main}
                activeDot={{ r: 8 }}
                isAnimationActive={false}
                dot={CustomDot}
                name='Avg. Ack'
              />
              <Line
                type='monotone'
                strokeWidth={2}
                dataKey='avgTimeToClose'
                isAnimationActive={false}
                strokeOpacity={props.loading ? 0.5 : 1}
                stroke={
                  theme.palette.mode === 'light'
                    ? theme.palette.secondary.dark
                    : theme.palette.secondary.light
                }
                activeDot={{ r: 8 }}
                dot={CustomDot}
                name='Avg. Close'
              />
            </LineChart>
          )}
        </AutoSizer>
      </Grid>
    </Grid>
  )
}
