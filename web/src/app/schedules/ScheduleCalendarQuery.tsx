import React from 'react'
import { gql, useQuery } from '@apollo/client'
import ScheduleCalendar from './ScheduleCalendar'
import { isWidthDown } from '@material-ui/core/withWidth/index'
import { getStartOfWeek, getEndOfWeek } from '../util/luxon-helpers'
import { DateTime } from 'luxon'
import useWidth from '../util/useWidth'
import { useURLParam } from '../actions/hooks'
import { Query } from '../../schema'
import { GenericError, ObjectNotFound } from '../error-pages'

const query = gql`
  query scheduleCalendarShifts(
    $id: ID!
    $start: ISOTimestamp!
    $end: ISOTimestamp!
  ) {
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
  onNewTempSched: () => void
  onEditTempSched: () => void
  onDeleteTempSched: () => void
}

function ScheduleCalendarQuery({
  scheduleID,
  ...other
}: ScheduleCalendarQueryProps): JSX.Element | null {
  const width = useWidth()
  const isMobile = isWidthDown('sm', width)

  const [weekly] = useURLParam<boolean>('weekly', false)
  const [start] = useURLParam(
    'start',
    weekly
      ? getStartOfWeek().toUTC().toISO()
      : DateTime.local().startOf('month').toUTC().toISO(),
  )

  const [queryStart, queryEnd] = weekly
    ? [
        getStartOfWeek(DateTime.fromISO(start)).toUTC().toISO(),
        getEndOfWeek(DateTime.fromISO(start)).toUTC().toISO(),
      ]
    : [
        getStartOfWeek(DateTime.fromISO(start).startOf('month'))
          .toUTC()
          .toISO(),
        getEndOfWeek(DateTime.fromISO(start).endOf('month')).toUTC().toISO(),
      ]

  const { data, error, loading } = useQuery<Query>(query, {
    variables: {
      id: scheduleID,
      start: queryStart,
      end: queryEnd,
    },
    skip: isMobile,
  })

  if (isMobile) return null
  if (error) return <GenericError error={error.message} />
  if (!loading && !data?.schedule?.id) return <ObjectNotFound type='schedule' />

  return (
    <ScheduleCalendar
      scheduleID={scheduleID}
      loading={loading && !data}
      shifts={data?.schedule?.shifts ?? []}
      temporarySchedules={data?.schedule?.temporarySchedules ?? []}
      {...other}
    />
  )
}

export default ScheduleCalendarQuery
