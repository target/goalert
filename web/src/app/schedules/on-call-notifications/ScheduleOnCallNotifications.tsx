import React from 'react'
import { gql, useQuery } from '@apollo/client'

import ScheduleOnCallNotificationCreateFab from './ScheduleOnCallNotificationCreateFab'
import { Schedule } from '../../../schema'
import { ObjectNotFound, GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import ScheduleOnCallNotificationsList from './ScheduleOnCallNotificationsList'

export const query = gql`
  query scheduleCalendarShifts($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
      onCallNotificationRules {
        id
        target {
          id
          type
          name
        }
        time
        weekdayFilter
      }
    }
  }
`

export const setMutation = gql`
  mutation ($input: SetScheduleOnCallNotificationRulesInput!) {
    setScheduleOnCallNotificationRules(input: $input)
  }
`

export const ScheduleContext = React.createContext<Schedule>({} as Schedule)

interface ScheduleOnCallNotificationsProps {
  scheduleID: string
}

export default function ScheduleOnCallNotifications(
  p: ScheduleOnCallNotificationsProps,
): JSX.Element {
  const { data, loading, error } = useQuery(query, {
    variables: {
      id: p.scheduleID,
    },
  })

  if (loading && !data) return <Spinner />
  if (data && !data.schedule) return <ObjectNotFound type='schedule' />
  if (error) return <GenericError error={error.message} />

  return (
    <ScheduleContext.Provider value={data.schedule}>
      <ScheduleOnCallNotificationsList />
      <ScheduleOnCallNotificationCreateFab />
    </ScheduleContext.Provider>
  )
}
