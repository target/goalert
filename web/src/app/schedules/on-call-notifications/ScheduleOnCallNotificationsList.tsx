import React, { useState } from 'react'
import { gql, useQuery } from '@apollo/client'
import { Grid } from '@material-ui/core'
import Avatar from '@material-ui/core/Avatar'

import QueryList from '../../lists/QueryList'
import OtherActions from '../../util/OtherActions'
import ScheduleOnCallNotificationCreateFab from './ScheduleOnCallNotificationCreateFab'
import { SlackBW } from '../../icons/components/Icons'
import { getRuleSummary, Rule } from './util'
import { useURLParam, useResetURLParams } from '../../actions/hooks'
import FilterContainer from '../../util/FilterContainer'
import { ScheduleTZFilter } from '../ScheduleTZFilter'
import ScheduleOnCallNotificationAction from './ScheduleOnCallNotificationAction'
import { Schedule } from '../../../schema'
import { ObjectNotFound, GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'

const query = gql`
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

export default function ScheduleOnCallNotificationsList(
  p: ScheduleOnCallNotificationsProps,
): JSX.Element {
  const [URLZone] = useURLParam('tz', 'local')
  const resetFilter = useResetURLParams('tz')
  const [editRule, setEditRule] = useState<Rule | null>(null)
  const [deleteRule, setDeleteRule] = useState<Rule | null>(null)

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
      <QueryList
        query={query}
        variables={{ id: p.scheduleID }}
        emptyMessage='No notification rules'
        noSearch
        path='data.onCallNotificationRules'
        headerNote={`Showing times in ${
          URLZone === 'local' ? 'local time' : URLZone
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
          subText: getRuleSummary(nr as Rule, data.schedule.timeZone, URLZone),
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
        handleOnCloseEdit={() => setEditRule(null)}
        handleOnCloseDelete={() => {
          setDeleteRule(null)
        }}
        editRule={editRule}
        deleteRule={deleteRule}
      />
      <ScheduleOnCallNotificationCreateFab />
    </ScheduleContext.Provider>
  )
}
