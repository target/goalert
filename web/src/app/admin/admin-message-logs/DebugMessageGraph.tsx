import React from 'react'
import { Card, CardContent, Grid, Paper, Typography } from '@mui/material'
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
import _ from 'lodash'

interface DebugMessageGraphProps {
  data: typeof LineChart.defaultProps['data']
  intervalType: string
}

export default function DebugMessageGraph(
  props: DebugMessageGraphProps,
): JSX.Element {
  const theme = useTheme()

  function getName(): string {
    switch (props.intervalType) {
      case 'daily':
        return 'Daily Message Count'
      default:
        return 'Message Count'
    }
  }

  return (
    <Card>
      <CardContent>
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
                  dataKey='count'
                  allowDecimals={false}
                  interval='preserveStart'
                  stroke={theme.palette.text.secondary}
                />
                <Tooltip
                  data-cy='message-count-tooltip'
                  cursor={{ fill: theme.palette.background.default }}
                  content={({ active, payload, label }) => {
                    if (!active || !payload?.length) return null

                    return (
                      <Paper variant='outlined' sx={{ p: 1 }}>
                        <Typography variant='body2'>{label}</Typography>
                        <Typography variant='body2'>
                          Count: {payload[0].payload.count}
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
                  name={getName()}
                  dataKey='count'
                />
              </LineChart>
            </ResponsiveContainer>
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  )
}
