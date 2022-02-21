import React from 'react'
import { Grid } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles/makeStyles'
import { Theme } from '@mui/material/styles'
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  Legend,
} from 'recharts'

interface AlertCountGraphProps {
  data: typeof BarChart.defaultProps['data']
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
  return (
    <Grid container className={classes.graphContent}>
      <Grid item xs={12}>
        <ResponsiveContainer width='100%' height='100%'>
          <BarChart
            width={730}
            height={250}
            data={props.data}
            margin={{
              top: 50,
              right: 30,
              bottom: 50,
            }}
          >
            <CartesianGrid strokeDasharray='4' vertical={false} />
            <XAxis dataKey='date' type='category' />
            <YAxis allowDecimals={false} dataKey='count' />
            <Tooltip
              labelFormatter={(label, props) => {
                return props?.length ? props[0].payload.label : label
              }}
            />
            <Legend />
            <Bar
              dataKey='count'
              fill='rgb(205, 24, 49)'
              className={classes.bar}
              name='Alert Count'
            />
          </BarChart>
        </ResponsiveContainer>
      </Grid>
    </Grid>
  )
}
