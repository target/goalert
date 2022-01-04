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
import { DateTime } from 'luxon'
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
    DateTime.fromISO(node.createdAt).toLocaleString(DateTime.DATE_SHORT),
  )

  const data = Object.entries(dateToAlerts).map(([date, alerts]) => ({
    name: date,
    count: alerts.length,
  }))

  return (
    <Grid container spacing={2} className={classes.gridContainer}>
      <Grid item xs={12}>
        <Card>
          <CardHeader
            component='h2'
            title='Daily alert counts over the last 28 days'
          />
          <CardContent>
            <Grid container>
              <Grid item xs={6}>
                <ServiceSelect
                  onChange={(v) => setServices(v)}
                  multiple
                  value={services}
                  label='Filter by Service'
                />
              </Grid>
              <Grid item xs={6}>
                {/*  date range filter spot holder */}
              </Grid>
            </Grid>
          </CardContent>
          <CardContent>
            <Grid container className={classes.graphContent}>
              <Grid item xs={12}>
                <ResponsiveContainer width='100%' height='100%'>
                  <BarChart width={730} height={250} data={data}>
                    <CartesianGrid strokeDasharray='3 3' />
                    <XAxis dataKey='name' />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Bar dataKey='count' fill='#82ca9d' />
                  </BarChart>
                </ResponsiveContainer>
              </Grid>
            </Grid>
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
