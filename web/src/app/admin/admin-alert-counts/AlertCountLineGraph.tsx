import React, { useState } from 'react'
import { Grid, Paper, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles/makeStyles'
import { Theme, useTheme } from '@mui/material/styles'
import {
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer,
  LineChart,
  Line,
  Legend,
  Tooltip,
} from 'recharts'

interface AlertCountLineGraphProps {
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

export default function AlertCountLineGraph(
  props: AlertCountLineGraphProps,
): JSX.Element {
  const [active, setActive] = useState('')
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
              allowDuplicatedCategory={false}
              stroke={theme.palette.text.secondary}
            />
            <YAxis
              type='number'
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
                          <Typography>{`${svc.name}: ${svc.value}`}</Typography>
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
                strokeWidth={active === series.serviceName ? 3 : 1}
                name={series.serviceName}
                stroke={
                  active === series.serviceName
                    ? theme.palette.primary.main
                    : theme.palette.mode === 'light'
                    ? theme.palette.secondary.light
                    : theme.palette.secondary.light
                }
                key={idx}
              />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </Grid>
    </Grid>
  )
}
