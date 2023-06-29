import React from 'react'
import { Grid, Paper, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles/makeStyles'
import { Theme, useTheme } from '@mui/material/styles'
import AutoSizer from 'react-virtualized-auto-sizer'
import Spinner from '../../loading/components/Spinner'
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  BarChart,
  Bar,
  Legend,
} from 'recharts'

interface AlertCountGraphProps {
  data: (typeof BarChart.defaultProps)['data']
  loading: boolean
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

export default function AlertCountGraph(
  props: AlertCountGraphProps,
): JSX.Element {
  const classes = useStyles()
  const theme = useTheme()
  return (
    <Grid container className={classes.graphContent}>
      <Grid item xs={12} data-cy='metrics-count-graph'>
        {props.loading && <Spinner />}
        <AutoSizer>
          {({ width, height }: { width: number; height: number }) => (
            <BarChart
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
                allowDecimals={false}
                dataKey='count'
                stroke={theme.palette.text.secondary}
              />
              <Tooltip
                data-cy='metrics-tooltip'
                cursor={{ fill: theme.palette.background.default }}
                content={({ active, payload, label }) => {
                  if (!active || !payload?.length) return null

                  const alertCountStr = `${payload[1].name}: ${
                    (payload[1].value as number) + (payload[0].value as number)
                  }`
                  const escalatedCountStr = `${payload[0].name}: ${payload[0].value}`
                  return (
                    <Paper variant='outlined' sx={{ p: 1 }}>
                      <Typography variant='body2'>{label}</Typography>
                      <Typography variant='body2'>{alertCountStr}</Typography>
                      <Typography variant='body2'>
                        {escalatedCountStr}
                      </Typography>
                    </Paper>
                  )
                }}
              />
              <Legend />
              <Bar
                dataKey='escalatedCount'
                stackId='a'
                fillOpacity={props.loading ? 0.5 : 1}
                fill={theme.palette.primary.main}
                className={classes.bar}
                name='Escalated'
              />
              <Bar
                stackId='a'
                dataKey='nonEscalatedCount'
                fillOpacity={props.loading ? 0.5 : 1}
                fill={
                  theme.palette.mode === 'light'
                    ? theme.palette.secondary.dark
                    : theme.palette.secondary.light
                }
                className={classes.bar}
                name='Alert Count'
              />
            </BarChart>
          )}
        </AutoSizer>
      </Grid>
    </Grid>
  )
}
