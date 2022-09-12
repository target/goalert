import React, { useState } from 'react'
import { Grid, FormControlLabel, Switch, Tooltip } from '@mui/material'
import QueryList from '../lists/QueryList'
import { gql } from '@apollo/client'
import { UserAvatar } from '../util/avatars'
import OtherActions from '../util/OtherActions'
import FilterContainer from '../util/FilterContainer'
import { UserSelect } from '../selection'
import { useURLParam, useResetURLParams } from '../actions'
import ScheduleOverrideCreateDialog from './ScheduleOverrideCreateDialog'
import ScheduleNewOverrideFAB from './ScheduleNewOverrideFAB'
import ScheduleOverrideDeleteDialog from './ScheduleOverrideDeleteDialog'
import { formatOverrideTime } from './util'
import ScheduleOverrideEditDialog from './ScheduleOverrideEditDialog'
import { useScheduleTZ } from './useScheduleTZ'
import { DateTime } from 'luxon'

// the query name `scheduleOverrides` is used for refetch queries
const query = gql`
  query scheduleOverrides($input: UserOverrideSearchOptions) {
    userOverrides(input: $input) {
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
  }
`

export default function ScheduleOverrideList({ scheduleID }) {
  const [editID, setEditID] = useState(null)
  const [deleteID, setDeleteID] = useState(null)
  const [create, setCreate] = useState(null)

  const [userFilter, setUserFilter] = useURLParam('userFilter', [])
  const [showPast, setShowPast] = useURLParam('showPast', false)
  const now = React.useMemo(() => new Date().toISOString(), [showPast])
  const resetFilter = useResetURLParams('userFilter', 'showPast', 'tz')

  const { zone, isLocalZone } = useScheduleTZ(scheduleID)

  const subText = (n) => {
    const tzTimeStr = formatOverrideTime(n.start, n.end, zone)
    const tzAbbr = DateTime.local({ zone }).toFormat('ZZZZ')
    const localTimeStr = formatOverrideTime(n.start, n.end, 'local')
    const localAbbr = DateTime.local({ zone: 'local' }).toFormat('ZZZZ')

    let tzSubText
    let localSubText
    if (n.addUser && n.removeUser) {
      // replace
      tzSubText = `Replaces ${n.removeUser.name} from ${tzTimeStr} ${tzAbbr}`
      localSubText = `Replaces ${n.removeUser.name} from ${localTimeStr} ${localAbbr}`
    } else if (n.addUser) {
      // add
      tzSubText = `Added from ${tzTimeStr} ${tzAbbr}`
      localSubText = `Added from ${localTimeStr} ${localAbbr}`
    } else {
      // remove
      tzSubText = `Removed from ${tzTimeStr} ${tzAbbr}`
      localSubText = `Removed from ${localTimeStr} ${localAbbr}`
    }

    return isLocalZone ? (
      <span>{localSubText}</span>
    ) : (
      <Tooltip
        title={localSubText}
        placement='bottom-start'
        PopperProps={{
          'aria-label': 'local-timezone-tooltip',
        }}
      >
        <span>{tzSubText}</span>
      </Tooltip>
    )
  }

  const hasUsers = Boolean(userFilter.length)
  const note = showPast
    ? `Showing all overrides${
        hasUsers ? ' for selected users' : ''
      } in ${zone}.`
    : `Showing active and future overrides${
        hasUsers ? ' for selected users' : ''
      } in ${zone}.`

  return (
    <React.Fragment>
      <ScheduleNewOverrideFAB onClick={(variant) => setCreate(variant)} />
      <QueryList
        headerNote={note}
        noSearch
        noPlaceholder
        query={query}
        mapDataNode={(n) => ({
          title: n.addUser ? n.addUser.name : n.removeUser.name,
          subText: subText(n),
          icon: (
            <UserAvatar userID={n.addUser ? n.addUser.id : n.removeUser.id} />
          ),
          action: (
            <OtherActions
              actions={[
                {
                  label: 'Edit',
                  onClick: () => setEditID(n.id),
                },
                {
                  label: 'Delete',
                  onClick: () => setDeleteID(n.id),
                },
              ]}
            />
          ),
        })}
        variables={{
          input: {
            scheduleID,
            start: showPast ? null : now,
            filterAnyUserID: userFilter,
          },
        }}
        headerAction={
          <FilterContainer onReset={() => resetFilter()}>
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={showPast}
                    onChange={(e) => setShowPast(e.target.checked)}
                    value='showPast'
                  />
                }
                label='Show past overrides'
              />
            </Grid>
            <Grid item xs={12}>
              <UserSelect
                label='Filter users...'
                multiple
                value={userFilter}
                onChange={(value) => setUserFilter(value)}
              />
            </Grid>
          </FilterContainer>
        }
      />
      {create && (
        <ScheduleOverrideCreateDialog
          scheduleID={scheduleID}
          variant={create}
          onClose={() => setCreate(null)}
        />
      )}
      {deleteID && (
        <ScheduleOverrideDeleteDialog
          overrideID={deleteID}
          onClose={() => setDeleteID(null)}
        />
      )}
      {editID && (
        <ScheduleOverrideEditDialog
          overrideID={editID}
          onClose={() => setEditID(null)}
        />
      )}
    </React.Fragment>
  )
}
