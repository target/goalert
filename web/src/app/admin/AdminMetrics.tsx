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

const useStyles = makeStyles<typeof theme>((theme) => ({
  gridContainer: {
    [theme.breakpoints.up('md')]: {
      justifyContent: 'center',
    },
  },
  groupTitle: {
    fontSize: '1.1rem',
  },
  saveDisabled: {
    color: 'rgba(255, 255, 255, 0.5)',
  },
  graphContent: {
    height: '500px',
  },
}))

export default function AdminMetrics(): JSX.Element {
  const classes = useStyles()
  const now = useMemo(() => DateTime.now(), [])
  const notCreatedBefore = now.minus({ weeks: 2 }).toISO()

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
      {/* <Grid item xs={12}>
        <Typography
          component='h2'
          variant='subtitle1'
          color='textSecondary'
          classes={{ subtitle1: classes.groupTitle }}
        >
          Alert Metrics
        </Typography>
      </Grid> */}
      <Grid item xs={12}>
        <Card>
          <CardHeader
            component='h3'
            title='Daily alert counts over the last 28 days'
          />
          <CardContent className={classes.graphContent}>
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
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}
