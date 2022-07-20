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

  const chooseColor = (idx: number): string => {
    switch (idx) {
      case 1:
        return '#3f8e98'
      case 2:
        return '#209e90'
      case 3:
        return '#41ab74'
      case 4:
        return '#79b34a'
      case 5:
        return '#b9b218'
      case 6:
        return '#ffa600'
      default:
        return '#607d8b'
    }
  }

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
                strokeWidth={active === series.serviceName ? 4 : 1}
                name={series.serviceName}
                stroke={chooseColor(idx)}
                key={idx}
              />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </Grid>
    </Grid>
  )
}
