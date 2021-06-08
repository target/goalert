import React from 'react'
import { gql } from '@apollo/client'
import QueryList from '../../lists/QueryList'
import ScheduleOnCallNotificationAction, {
  Rule,
} from './ScheduleOnCallNotificationAction'
import ScheduleOnCallNotificationCreateFab from './ScheduleOnCallNotificationCreateFab'
import { SlackBW } from '../../icons/components/Icons'

import Avatar from '@material-ui/core/Avatar'
import { getSelectedDays } from './ScheduleOnCallNotificationFormDialog'

interface ScheduleOnCallNotificationsProps {
  scheduleID: string
}

export const query = gql`
  query scheduleCalendarShifts($id: ID!) {
    schedule(id: $id) {
      id
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

function subText(rule: Rule): string {
  if (rule.time && rule.weekdayFilter) {
    const days = getSelectedDays(rule.weekdayFilter).map((d) => d.label + 's')
    const lastDay = days.length > 1 ? days.pop() : ''

    return (
      'Notifies on ' +
      days.join(', ') +
      (lastDay && ' and ' + lastDay) +
      ' at ' +
      rule.time
    )
  }

  return 'Notifies when on-call hands off'
}

export default function ScheduleOnCallNotificationsList(
  p: ScheduleOnCallNotificationsProps,
): JSX.Element {
  return (
    <React.Fragment>
      <QueryList
        query={query}
        variables={{ id: p.scheduleID }}
        headerNote='Configure notifications for on-call updates'
        emptyMessage='No notification rules'
        noSearch
        path='onCallNotificationRules'
        mapDataNode={(nr) => ({
          id: nr.id,
          icon: (
            <Avatar>
              <SlackBW />
            </Avatar>
          ),
          title: nr.target.name,
          subText: subText(nr as Rule),
          action: (
            <ScheduleOnCallNotificationAction
              rule={nr as Rule}
              scheduleID={p.scheduleID}
            />
          ),
        })}
      />
      <ScheduleOnCallNotificationCreateFab scheduleID={p.scheduleID} />
    </React.Fragment>
  )
}
