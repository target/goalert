import React from 'react'
import { gql, useQuery } from '@apollo/client'
import ScheduleCalendar from './ScheduleCalendar'
import { isWidthDown } from '@material-ui/core/withWidth/index'
import { getStartOfWeek } from '../util/luxon-helpers'
import { DateTime } from 'luxon'
import useWidth from '../util/useWidth'
import { useURLParam } from '../actions/hooks'
import { Query } from '../../schema'
import { GenericError, ObjectNotFound } from '../error-pages'
import Spinner from '../loading/components/Spinner'

const query = gql`
  query scheduleCalendarShifts(
    $id: ID!
    $start: ISOTimestamp!
    $end: ISOTimestamp!
  ) {
    schedule(id: $id) {
      id
      shifts(start: $start, end: $end) {
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
`

interface ScheduleCalendarQueryProps {
  scheduleID: string
}

function ScheduleCalendarQuery({
  scheduleID,
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

  const end = weekly
    ? DateTime.fromISO(start).plus({ weeks: 1 }).toUTC().toISO()
    : DateTime.fromISO(start).endOf('month').toUTC().toISO()

  const { data, error, loading } = useQuery<Query>(query, {
    variables: {
      id: scheduleID,
      start,
      end,
    },
    skip: isMobile,
  })

  if (isMobile) return null
  if (loading && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />
  if (!data?.schedule?.id) return <ObjectNotFound type='schedule' />

  return (
    <ScheduleCalendar
      scheduleID={scheduleID}
      shifts={data?.schedule?.shifts ?? []}
    />
  )
}

export default ScheduleCalendarQuery
