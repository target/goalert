import React from 'react'
import { gql } from '@apollo/client'
import QueryList from '../../lists/QueryList'
import ScheduleOnCallNotificationAction from './ScheduleOnCallNotificationAction'

interface ScheduleOnCallNotificationsProps {
  scheduleID: string
}

export const query = gql`
  query scheduleCalendarShifts($id: ID!) {
    schedule(id: $id) {
      id
      notificationRules {
        id
        channel
        filter
        time
        allChanges
      }
    }
  }
`

export const setMutation = gql`
  mutation ($input: [SetScheduleNotificationsInput!]) {
    setScheduleNotifications(input: $input)
  }
`

export default function ScheduleOnCallNotifications(
  p: ScheduleOnCallNotificationsProps,
): JSX.Element {
  return (
    <QueryList
      query={query}
      variables={{ id: p.scheduleID }}
      headerNote='Configure notifications for on-call updates'
      noSearch
      mapDataNode={(n) => ({
        id: n.id,
        title: n.channel,
        action: (
          <ScheduleOnCallNotificationAction
            id={n.id}
            scheduleID={p.scheduleID}
          />
        ),
      })}
    />
  )
}
