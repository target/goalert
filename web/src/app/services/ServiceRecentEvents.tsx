import React, { useState } from 'react'
import {
  Card,
  CardHeader,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  SelectChangeEvent,
} from '@mui/material'
import { gql, useQuery } from 'urql'
import { GenericError } from '../error-pages'
import CompList from '../lists/CompList'
import { CompListItemNav } from '../lists/CompListItems'
import { AlertLogEntry } from '../../schema'
import { DateTime } from 'luxon'
import { Time } from '../util/Time'

const query = gql`
  query ServiceRecentEvents($serviceID: ID!, $since: ISOTimestamp) {
    service(id: $serviceID) {
      id
      recentEvents(input: { since: $since, limit: 5 }) {
        nodes {
          id
          alertID
          timestamp
          message
        }
      }
    }
  }
`

export interface ServiceRecentEventsProps {
  serviceID: string
}

type TimeRange = '1h' | '1d' | '1w'

const timeRangeOptions = [
  { value: '1h' as TimeRange, label: '1 hour' },
  { value: '1d' as TimeRange, label: '1 day' },
  { value: '1w' as TimeRange, label: '1 week' },
]

function getTimestamp(range: TimeRange): string {
  const now = DateTime.now().startOf('minute')
  switch (range) {
    case '1h':
      return now.plus({ hour: -1 }).toISO()
    case '1d':
      return now.plus({ day: -1 }).toISO()
    case '1w':
      return now.plus({ week: -1 }).toISO()
    default:
      throw new Error(`Unknown time range: ${range}`)
  }
}

function formatEventMessage(entry: AlertLogEntry): string {
  // Extract the alert ID and format a simple message
  return `Alert ${entry.alertID}: ${entry.message}`
}

const noSuspense = {
  suspense: false,
}
export default function ServiceRecentEvents({
  serviceID,
}: ServiceRecentEventsProps): React.JSX.Element {
  const [timeRange, setTimeRange] = useState<TimeRange>('1h')

  const [{ data, error, fetching }] = useQuery({
    query,
    variables: {
      serviceID,
      since: getTimestamp(timeRange),
    },
    context: noSuspense,
  })

  const handleTimeRangeChange = (event: SelectChangeEvent<TimeRange>): void => {
    setTimeRange(event.target.value as TimeRange)
  }

  if (error) {
    return <GenericError error={error.message} />
  }

  const events = data?.service?.recentEvents?.nodes || []

  return (
    <Card data-testid='service-recent-events'>
      <CardHeader
        title='Most recent events'
        action={
          <FormControl size='small' sx={{ minWidth: 120 }}>
            <InputLabel>Time Range</InputLabel>
            <Select
              value={timeRange}
              label='Time Range'
              onChange={handleTimeRangeChange}
            >
              {timeRangeOptions.map((option) => (
                <MenuItem key={option.value} value={option.value}>
                  {option.label}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        }
      />
      <CompList
        data-cy='service-recent-events'
        emptyMessage={
          fetching
            ? 'Loading events...'
            : 'No recent events in the selected time range'
        }
      >
        {events.map((event: AlertLogEntry) => (
          <CompListItemNav
            key={event.id}
            title={formatEventMessage(event)}
            subText={<Time time={event.timestamp} format='relative' />}
            url={`/alerts/${event.alertID}`}
          />
        ))}
      </CompList>
    </Card>
  )
}
