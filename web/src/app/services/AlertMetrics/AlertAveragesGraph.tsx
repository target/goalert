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
              content={(props) => {
                const ackAvg = props.payload?.length
                  ? `${props.payload[0].name}: ${props.payload[0].payload.formattedAckLabel}`
                  : ''
                const closeAvg = props.payload?.length
                  ? `${props.payload[1].name}: ${props.payload[1].payload.formattedCloseLabel}`
                  : ''
                return (
                  <Paper variant='outlined' sx={{ p: 1 }}>
                    <Typography variant='body2'>{props.label}</Typography>
                    <Typography variant='body2'>{ackAvg}</Typography>
                    <Typography variant='body2'>{closeAvg}</Typography>
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
              dot={CustomDot}
              name='Average Time To Acknowledge'
            />
            <Line
              type='monotone'
              strokeWidth={2}
              dataKey='avgTimeToClose'
              stroke={
                theme.palette.mode === 'light'
                  ? theme.palette.secondary.dark
                  : theme.palette.secondary.light
              }
              activeDot={{ r: 8 }}
              dot={CustomDot}
              name='Average Time To Close'
            />
          </LineChart>
        </ResponsiveContainer>
      </Grid>
    </Grid>
  )
}
