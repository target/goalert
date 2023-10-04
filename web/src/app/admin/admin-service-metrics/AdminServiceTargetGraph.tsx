import React from 'react'
import { TargetMetrics } from './useServiceMetrics'
import {
  BarChart,
  CartesianGrid,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  Bar,
} from 'recharts'
import { useTheme, Theme } from '@mui/material/styles'
import AutoSizer from 'react-virtualized-auto-sizer'
import { Grid } from '@mui/material'
import { makeStyles } from '@mui/styles'

interface AdminServiceTargetGraphProps {
  metrics: TargetMetrics
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
export default function AdminServiceTargetGraph(
  props: AdminServiceTargetGraphProps,
): JSX.Element {
  const theme = useTheme()
  const classes = useStyles()
  const { metrics, loading } = props
  let targetMetrics = [] as { type: string; count: number }[]
  if (metrics) {
    targetMetrics = Object.entries(metrics).map(([type, count]) => ({
      type,
      count,
    }))
  }

  return (
    <Grid container className={classes.graphContent}>
      <Grid item xs={12} data-cy='alert-count-graph'>
        <AutoSizer>
          {({ width, height }: { width: number; height: number }) => (
            <BarChart
              width={width}
              height={height}
              data={targetMetrics}
              margin={{
                top: 50,
                right: 30,
                bottom: 50,
              }}
            >
              <CartesianGrid strokeDasharray='3 3' />
              <XAxis dataKey='type' />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar
                dataKey='count'
                fill={
                  loading
                    ? theme.palette.secondary.light
                    : theme.palette.primary.main
                }
              />
            </BarChart>
          )}
        </AutoSizer>
      </Grid>
    </Grid>
  )
}
