import React from 'react'
import { gql, useQuery } from 'urql'
import { getStartOfWeek, getEndOfWeek } from '../util/luxon-helpers'
import { DateTime } from 'luxon'
import { useIsWidthDown } from '../util/useWidth'
import { GenericError } from '../error-pages'
import { useCalendarNavigation } from '../util/calendar/hooks'
import Calendar from '../util/calendar/Calendar'

const query = gql`
  query userCalendarShifts(
    $id: ID!
    $start: ISOTimestamp!
    $end: ISOTimestamp!
  ) {
    userShifts(input: { id: $id, start: $start, end: $end }) {
      start
      end
      truncated
      user {
        id
        name
      }
    }
  }
`

interface UserShiftsCalendarProps {
  userID: string
}

export default function UserShiftsCalendar({
  userID,
}: UserShiftsCalendarProps): JSX.Element | null {
  const isMobile = useIsWidthDown('md')
  const { weekly, start } = useCalendarNavigation()

  const [queryStart, queryEnd] = weekly
    ? [
        getStartOfWeek(DateTime.fromISO(start)).toISO(),
        getEndOfWeek(DateTime.fromISO(start)).toISO(),
      ]
    : [
        getStartOfWeek(DateTime.fromISO(start).startOf('month')).toISO(),
        getEndOfWeek(DateTime.fromISO(start).endOf('month')).toISO(),
      ]

  const [{ data, error, fetching }] = useQuery({
    query,
    variables: {
      id: userID,
      start: queryStart,
      end: queryEnd,
    },
    pause: isMobile,
  })

  if (isMobile) return null
  if (error) return <GenericError error={error.message} />

  return <Calendar loading={fetching} shifts={data?.userShifts} />
}
