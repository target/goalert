import React, { useState, useMemo } from 'react'
import p from 'prop-types'
import { Grid, FormControlLabel, Switch } from '@material-ui/core'
import { gql } from '@apollo/client'

import { UserAvatar } from '../util/avatars'
import OtherActions from '../util/OtherActions'
import FilterContainer from '../util/FilterContainer'
import { UserSelect } from '../selection'
import { useURLParam, useResetURLParams } from '../actions'
import { ScheduleTZFilter } from './ScheduleTZFilter'
import ScheduleOverrideCreateDialog from './ScheduleOverrideCreateDialog'
import ScheduleNewOverrideFAB from './ScheduleNewOverrideFAB'
import ScheduleOverrideDeleteDialog from './ScheduleOverrideDeleteDialog'
import { formatOverrideTime } from './util'
import ScheduleOverrideEditDialog from './ScheduleOverrideEditDialog'
import { usePaginatedQuery } from '../lists/usePaginatedQuery'
import { PaginatedList } from '../lists/PaginatedList'
import { useSelector } from 'react-redux'
import { urlKeySelector } from '../selectors'

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

export default function ScheduleOverrideList(props) {
  const [editID, setEditID] = useState(null)
  const [deleteID, setDeleteID] = useState(null)
  const [create, setCreate] = useState(null)
  const urlKey = useSelector(urlKeySelector)
  const [userFilter, setUserFilter] = useURLParam('userFilter', [])
  const [showPast, setShowPast] = useURLParam('showPast', false)
  const [zone] = useURLParam('tz', 'local')
  const resetFilter = useResetURLParams('userFilter', 'showPast', 'tz')

  const now = useMemo(() => new Date().toISOString(), [showPast])

  const subText = (n) => {
    const timeStr = formatOverrideTime(n.start, n.end, zone)
    if (n.addUser && n.removeUser) {
      // replace
      return `Replaces ${n.removeUser.name} from ${timeStr}`
    }
    if (n.addUser) {
      // add
      return `Added from ${timeStr}`
    }
    // remove
    return `Removed from ${timeStr}`
  }

  const { q, loadMore } = usePaginatedQuery(
    query,
    {
      input: {
        scheduleID: props.scheduleID,
        start: showPast ? null : now,
        filterAnyUserID: userFilter,
      },
    },
    false,
  )

  const items =
    q.data?.data?.nodes?.map((n) => ({
      title: n.addUser ? n.addUser.name : n.removeUser.name,
      subText: subText(n),
      icon: <UserAvatar userID={n.addUser ? n.addUser.id : n.removeUser.id} />,
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
    })) ?? []

  const zoneText = zone === 'local' ? 'local time' : zone
  const hasUsers = Boolean(userFilter.length)
  const headerNote = showPast
    ? `Showing all overrides${
        hasUsers ? ' for selected users' : ''
      } in ${zoneText}.`
    : `Showing active and future overrides${
        hasUsers ? ' for selected users' : ''
      } in ${zoneText}.`

  return (
    <React.Fragment>
      <ScheduleNewOverrideFAB onClick={(variant) => setCreate(variant)} />
      <PaginatedList
        key={urlKey}
        noPlaceholder
        items={items}
        isLoading={!q.data && q.loading}
        loadMore={loadMore}
        headerNote={headerNote}
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
              <ScheduleTZFilter scheduleID={props.scheduleID} />
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
          scheduleID={props.scheduleID}
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

ScheduleOverrideList.propTypes = {
  scheduleID: p.string.isRequired,
}
