import React, { useState } from 'react'
import { gql } from '@apollo/client'
import QueryList from '../../lists/QueryList'
import OtherActions from '../../util/OtherActions'
import ScheduleOnCallNotificationCreateFab from './ScheduleOnCallNotificationCreateFab'
import { SlackBW } from '../../icons/components/Icons'
import { Grid } from '@material-ui/core'
import Avatar from '@material-ui/core/Avatar'
import { getDayNames, Rule } from './util'
import { useURLParam, useResetURLParams } from '../../actions/hooks'
import FilterContainer from '../../util/FilterContainer'
import { ScheduleTZFilter } from '../ScheduleTZFilter'
import { DateTime } from 'luxon'
import ScheduleOnCallNotificationAction from './ScheduleOnCallNotificationAction'

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

function subText(rule: Rule, zone: string): string {
  if (rule.time && rule.weekdayFilter) {
    const timeStr = DateTime.fromFormat(rule.time, 'HH:mm', {
      zone,
    }).toISOTime()
    return `Notifies ${getDayNames(rule.weekdayFilter)} at ${DateTime.fromISO(
      timeStr,
    ).toLocaleString(DateTime.TIME_SIMPLE)}
    `
  }

  return 'Notifies when on-call hands off'
}

export default function ScheduleOnCallNotificationsList(
  p: ScheduleOnCallNotificationsProps,
): JSX.Element {
  const [zone] = useURLParam('tz', 'local')
  const resetFilter = useResetURLParams('tz')
  const [editRule, setEditRule] = useState<Rule | undefined>(undefined)
  const [deleteRule, setDeleteRule] = useState<Rule | undefined>(undefined)

  return (
    <React.Fragment>
      <QueryList
        query={query}
        variables={{ id: p.scheduleID }}
        emptyMessage='No notification rules'
        noSearch
        path='data.onCallNotificationRules'
        headerNote={`Showing times in ${
          zone === 'local' ? 'local time' : zone
        }.`}
        headerAction={
          <FilterContainer onReset={() => resetFilter()}>
            <Grid item xs={12}>
              <ScheduleTZFilter scheduleID={p.scheduleID} />
            </Grid>
          </FilterContainer>
        }
        mapDataNode={(nr) => ({
          id: nr.id,
          icon: (
            <Avatar>
              <SlackBW />
            </Avatar>
          ),
          title: nr.target.name,
          subText: subText(nr as Rule, zone),
          action: (
            <OtherActions
              actions={[
                { label: 'Edit', onClick: () => setEditRule(nr as Rule) },
                { label: 'Delete', onClick: () => setDeleteRule(nr as Rule) },
              ]}
            />
          ),
        })}
      />
      <ScheduleOnCallNotificationAction
        scheduleID={p.scheduleID}
        handleOnCloseEdit={() => setEditRule(undefined)}
        handleOnCloseDelete={() => {
          setDeleteRule(undefined)
        }}
        editRule={editRule}
        deleteRule={deleteRule}
      />
      <ScheduleOnCallNotificationCreateFab scheduleID={p.scheduleID} />
    </React.Fragment>
  )
}
