import React from 'react'
import { gql, useQuery } from 'urql'
import { getStartOfWeek, getEndOfWeek } from '../util/luxon-helpers'
import { DateTime } from 'luxon'
import { GenericError } from '../error-pages'
import { useCalendarNavigation } from '../util/calendar/hooks'
import Calendar, { Shift } from '../util/calendar/Calendar'
import { OnCallShift, Schedule } from '../../schema'

const query = gql`
  query user($id: ID!, $start: ISOTimestamp!, $end: ISOTimestamp!) {
    user(id: $id) {
      id
      assignedSchedules {
        id
        name
        shifts(start: $start, end: $end, userIDs: [$id]) {
          start
          end
        }
      }
    }
  }
`

interface UserShiftsCalendarProps {
  userID: string
}

export default function UserShiftsCalendar({
  userID,
}: UserShiftsCalendarProps): React.JSX.Element | null {
  const { weekly, start } = useCalendarNavigation()

  const queryStart = weekly
    ? getStartOfWeek(DateTime.fromISO(start)).toISO()
    : getStartOfWeek(DateTime.fromISO(start).startOf('month')).toISO()

  const queryEnd = weekly
    ? getEndOfWeek(DateTime.fromISO(start)).toISO()
    : getEndOfWeek(DateTime.fromISO(start).endOf('month')).toISO()

  const [{ data, error }] = useQuery({
    query,
    variables: {
      id: userID,
      start: queryStart,
      end: queryEnd,
    },
  })

  if (error) return <GenericError error={error.message} />

  function makeCalendarShifts(): OnCallShift[] {
    const assignedSchedules: Schedule[] = data?.user?.assignedSchedules ?? []

    const s: Shift[] = []
    assignedSchedules.forEach((a) => {
      a.shifts.forEach((shift) => {
        s.push({
          userID: shift.userID,
          start: shift.start,
          end: shift.end,
          truncated: shift.truncated,
          targetName: a.name,
          targetID: a.id,
        })
      })
    })

    return s
  }

  return (
    <Calendar loading={false} shifts={makeCalendarShifts()} showScheduleLink />
  )
}
