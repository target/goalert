import React, { useMemo } from 'react'
import { Card, CardContent, CardHeader, Grid } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles/makeStyles'
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
import { useQuery, gql } from '@apollo/client'
import { theme } from '../mui'
import { DateTime, Interval } from 'luxon'
import _ from 'lodash'
import { ServiceSelect } from '../selection'
import { useURLParam } from '../actions/hooks'

const useStyles = makeStyles<typeof theme>((theme) => ({
  gridContainer: {
    [theme.breakpoints.up('md')]: {
      justifyContent: 'center',
    },
  },
  groupTitle: {
    fontSize: '1.1rem',
  },
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

export default function AdminMetrics(): JSX.Element {
  const classes = useStyles()
  const [services, setServices] = useURLParam<string[]>('services', [])
  const now = useMemo(() => DateTime.now(), [])
  const notCreatedBefore = now.minus({ weeks: 4 }).toISO()

  const query = gql`
    query alerts($input: AlertSearchOptions!) {
      alerts(input: $input) {
        nodes {
          id
          createdAt
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  `

  const q = useQuery(query, {
    variables: {
      input: {
        filterByServiceID: services.length ? services : null,
        first: 100,
        notCreatedBefore,
      },
    },
  })

  const dateToAlerts = _.groupBy(q?.data?.alerts?.nodes ?? [], (node) =>
    DateTime.fromISO(node.createdAt).toLocaleString({
      month: 'short',
      day: 'numeric',
    }),
  )

  const data = Interval.fromDateTimes(
    DateTime.fromISO(notCreatedBefore).startOf('day'),
    now.endOf('day'),
  )
    .splitBy({ days: 1 })
    .map((day) => {
      let alertCount = 0
      const date = day.start.toLocaleString({ month: 'short', day: 'numeric' })

      if (dateToAlerts[date]) {
        alertCount = dateToAlerts[date].length
      }

      return {
        date: date,
        count: alertCount,
      }
    })

  return (
    <Grid container spacing={2} className={classes.gridContainer}>
      <Grid item xs={12}>
        <Card>
          <CardHeader
            component='h3'
            title='Daily alert counts over the last 28 days'
          />
          <CardContent>
            <Grid container>
              <Grid item xs={6}>
                <ServiceSelect
                  onChange={(e) => setServices(e)}
                  multiple
                  value={services}
                />
              </Grid>
              <Grid item xs={6}>
                {/*  date range filter spot holder */}
              </Grid>
            </Grid>
            <Grid container className={classes.graphContent}>
              <ResponsiveContainer>
                <BarChart
                  width={730}
                  height={250}
                  data={data}
                  margin={{
                    top: 50,
                    right: 30,
                  }}
                >
                  <CartesianGrid strokeDasharray='4' vertical={false} />
                  <XAxis dataKey='date' type='category' />
                  <YAxis allowDecimals={false} dataKey='count' />
                  <Tooltip />
                  <Legend />
                  <Bar
                    dataKey='count'
                    fill={theme.palette.primary.main}
                    className={classes.bar}
                    name='Alert Count'
                  />
                </BarChart>
              </ResponsiveContainer>
            </Grid>
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
