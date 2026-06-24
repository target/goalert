import React from 'react'
import { gql, useQuery } from 'urql'
import { getStartOfWeek, getEndOfWeek } from '../util/luxon-helpers'
import { DateTime } from 'luxon'
import { useIsWidthDown } from '../util/useWidth'
import { GenericError, ObjectNotFound } from '../error-pages'
import { useCalendarNavigation } from '../util/calendar/hooks'
import Calendar from '../util/calendar/Calendar'

const query = gql`
  query scheduleCalendarShifts(
    $id: ID!
    $start: ISOTimestamp!
    $end: ISOTimestamp!
  ) {
    userOverrides(
      input: { scheduleID: $id, start: $start, end: $end, first: 149 }
    ) {
      # todo - make query expandable to handle >149 overrides
      nodes {
        id
        start
        end
        addUser {
          id
          name
        }
        removeUser {
          id
          name
        }
      }

      pageInfo {
        hasNextPage
        endCursor
      }
    }

    schedule(id: $id) {
      id
      shifts(start: $start, end: $end) {
        userID
        user {
          id
          name
        }
        start
        end
        truncated
      }

      temporarySchedules {
        start
        end
        shifts {
          userID
          user {
            id
            name
          }
          start
          end
          truncated
        }
      }
    }
  }
`

interface ScheduleCalendarQueryProps {
  scheduleID: string
}

function ScheduleCalendarQuery({
  scheduleID,
}: ScheduleCalendarQueryProps): React.JSX.Element | null {
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
      id: scheduleID,
      start: queryStart,
      end: queryEnd,
    },
    pause: isMobile,
  })

  if (isMobile) return null
  if (error) return <GenericError error={error.message} />
  if (!fetching && !data?.schedule?.id)
    return <ObjectNotFound type='schedule' />

  return (
    <Calendar
      scheduleID={scheduleID}
      loading={fetching}
      shifts={data?.schedule?.shifts ?? []}
      temporarySchedules={data?.schedule?.temporarySchedules ?? []}
      overrides={data?.userOverrides?.nodes ?? []}
    />
  )
}

export default ScheduleCalendarQuery
