import React, { useMemo } from 'react'
import {
  Card,
  CardContent,
  CardHeader,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Select,
  SelectChangeEvent,
} from '@mui/material'
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

const useStyles = makeStyles<typeof theme>((theme) => ({
  gridContainer: {
    [theme.breakpoints.up('md')]: {
      justifyContent: 'center',
    },
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
  const now = useMemo(() => DateTime.now(), [])
  const [minDate, maxDate] = [
    useMemo(() => now.minus({ weeks: 4 }).startOf('day'), [now]),
    now,
  ]
  const [services, setServices] = useURLParam<string[]>('services', [])
  const [_since, setSince] = useURLParam('since', minDate.toISO())
  const since = DateTime.max(DateTime.fromISO(_since), minDate).toISO()
  const until = maxDate

  const q = useQuery(query, {
    variables: {
      input: {
        filterByServiceID: services.length ? services : null,
        first: 100,
        createdBefore: until,
        notCreatedBefore: since,
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
    DateTime.fromISO(since).startOf('day'),
    until.endOf('day'),
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

  const handleDateRangeChange = (e: SelectChangeEvent<number>): void => {
    const weeks = e?.target?.value as number
    setSince(now.minus({ weeks }).startOf('day').toISO())
  }

  return (
    <Grid container spacing={2} className={classes.gridContainer}>
      <Grid item xs={12}>
        <Card>
          <CardHeader
            component='h2'
            title='Daily alert counts over the last 28 days'
          />
          <CardContent>
            <Grid container justifyContent='space-around'>
              <Grid item xs={5}>
                <ServiceSelect
                  onChange={(v) => setServices(v)}
                  multiple
                  value={services}
                  label='Filter by Service'
                />
              </Grid>
              <Grid item xs={5}>
                <FormControl sx={{ width: '100%' }}>
                  <InputLabel id='demo-simple-select-helper-label'>
                    Date Range
                  </InputLabel>
                  <Select
                    fullWidth
                    labelId='demo-simple-select-helper-label'
                    id='demo-simple-select-helper'
                    value={Math.floor(
                      -DateTime.fromISO(since).diffNow('weeks').weeks,
                    )}
                    label='Date Range'
                    name='date-range'
                    onChange={handleDateRangeChange}
                  >
                    <MenuItem value={1}>Past week</MenuItem>
                    <MenuItem value={2}>Past 2 weeks</MenuItem>
                    <MenuItem value={3}>Past 3 weeks</MenuItem>
                    <MenuItem value={4}>Past 4 weeks</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
            </Grid>
          </CardContent>
          <CardContent>
            <Grid container className={classes.graphContent}>
              <Grid item xs={12}>
                <ResponsiveContainer width='100%' height='100%'>
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
                      fill='rgb(205, 24, 49)'
                      className={classes.bar}
                      name='Alert Count'
                    />
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
