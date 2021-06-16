import React, { useContext, useState } from 'react'
import { Grid } from '@material-ui/core'
import Avatar from '@material-ui/core/Avatar'

import QueryList from '../../lists/QueryList'
import OtherActions from '../../util/OtherActions'
import { SlackBW } from '../../icons/components/Icons'
import { getRuleSummary, Rule } from './util'
import { useURLParam, useResetURLParams } from '../../actions/hooks'
import FilterContainer from '../../util/FilterContainer'
import { ScheduleTZFilter } from '../ScheduleTZFilter'
import { query, ScheduleContext } from './ScheduleOnCallNotifications'
import ScheduleOnCallNotificationsFormDialog from './ScheduleOnCallNotificationsFormDialog'
import ScheduleOnCallNotificationsDeleteDialog from './ScheduleOnCallNotificationsDeleteDialog'

export default function ScheduleOnCallNotificationsList(): JSX.Element {
  const [displayZone] = useURLParam('tz', 'local')
  const resetFilter = useResetURLParams('tz')
  const schedCtx = useContext(ScheduleContext)
  const [editRule, setEditRule] = useState<Rule | null>(null)
  const [deleteRule, setDeleteRule] = useState<Rule | null>(null)

  return (
    <React.Fragment>
      <QueryList
        query={query}
        variables={{ id: schedCtx.id }}
        emptyMessage='No notification rules'
        noSearch
        path='data.onCallNotificationRules'
        headerNote={`Showing times in ${
          displayZone === 'local' ? 'local time' : displayZone
        }.`}
        headerAction={
          <FilterContainer onReset={() => resetFilter()}>
            <Grid item xs={12}>
              <ScheduleTZFilter scheduleID={schedCtx.id} />
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
          subText: getRuleSummary(nr as Rule, schedCtx.timeZone, displayZone),
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
      {editRule && (
        <ScheduleOnCallNotificationsFormDialog
          rule={editRule}
          onClose={() => setEditRule(null)}
        />
      )}
      {deleteRule && (
        <ScheduleOnCallNotificationsDeleteDialog
          rule={deleteRule}
          onClose={() => setDeleteRule(null)}
        />
      )}
    </React.Fragment>
  )
}
