import React from 'react'
import { Grid, Paper, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles/makeStyles'
import { Theme, useTheme } from '@mui/material/styles'
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  LineChart,
  Line,
  Legend,
  DotProps,
} from 'recharts'

interface CustomDotProps extends DotProps {
  dataKey: string
  payload: { date: string }
}

const CustomDot = (props: CustomDotProps): JSX.Element => {
  const { cy, cx, fill, r, stroke, strokeWidth, dataKey, payload } = props
  return (
    <circle
      cy={cy}
      cx={cx}
      fill={fill}
      r={r}
      stroke={stroke}
      strokeWidth={strokeWidth}
      key={dataKey + '-' + payload.date}
      data-cy={dataKey + '-' + payload.date}
    />
  )
}

interface AlertAveragesGraphProps {
  data: typeof LineChart.defaultProps['data']
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

export default function AlertAveragesGraph(
  props: AlertAveragesGraphProps,
): JSX.Element {
  const classes = useStyles()
  const theme = useTheme()
  return (
    <Grid container className={classes.graphContent}>
      <Grid item xs={12} data-cy='metrics-averages-graph'>
        <ResponsiveContainer width='100%' height='100%'>
          <LineChart
            width={730}
            height={250}
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
        </ResponsiveContainer>
      </Grid>
    </Grid>
  )
}
