import React from 'react'
import { Grid, Paper, Typography } from '@mui/material'
import { useTheme } from '@mui/material/styles'
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  LineChart,
  Line,
  Legend,
} from 'recharts'

interface DebugMessageGraphProps {
  data: typeof LineChart.defaultProps['data']
}

export default function DebugMessageGraph(
  props: DebugMessageGraphProps,
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
              data-cy='message-count-tooltip'
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
              dataKey='ivl-message-count'
              strokeWidth={2}
              stroke={theme.palette.primary.main}
              activeDot={{ r: 8 }}
              isAnimationActive={false}
              dot={(props) => <circle {...props} />}
              name='Count'
            />
          </LineChart>
        </ResponsiveContainer>
      </Grid>
    </Grid>
  )
}
