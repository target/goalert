import React, { useCallback, useState } from 'react'
import { Button, Grid, FormControlLabel, Switch } from '@mui/material'
import { GroupAdd } from '@mui/icons-material'
import { gql } from '@apollo/client'
import QueryList from '../lists/QueryList'
import { UserAvatar } from '../util/avatars'
import OtherActions from '../util/OtherActions'
import FilterContainer from '../util/FilterContainer'
import { UserSelect } from '../selection'
import { useURLParam, useResetURLParams } from '../actions'
import ScheduleOverrideDeleteDialog from './ScheduleOverrideDeleteDialog'
import ScheduleOverrideEditDialog from './ScheduleOverrideEditDialog'
import { useScheduleTZ } from './useScheduleTZ'
import { useIsWidthDown } from '../util/useWidth'
import { OverrideDialogContext } from './ScheduleDetails'
import TempSchedDialog from './temp-sched/TempSchedDialog'
import ScheduleOverrideDialog from './ScheduleOverrideDialog'
import CreateFAB from '../lists/CreateFAB'
import { Time } from '../util/Time'

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
  const isMobile = useIsWidthDown('md')

  const [editID, setEditID] = useState(null)
  const [deleteID, setDeleteID] = useState(null)

  const [userFilter, setUserFilter] = useURLParam('userFilter', [])
  const [showPast, setShowPast] = useURLParam('showPast', false)
  const now = React.useMemo(() => new Date().toISOString(), [showPast])
  const resetFilter = useResetURLParams('userFilter', 'showPast', 'tz')

  const [overrideDialog, setOverrideDialog] = useState(null)
  const [configTempSchedule, setConfigTempSchedule] = useState(null)
  const onNewTempSched = useCallback(() => setConfigTempSchedule({}), [])

  const { zone } = useScheduleTZ(scheduleID)

  const subText = (n) => {
    let prefix
    if (n.addUser && n.removeUser) prefix = 'Replaces'
    else if (n.addUser) prefix = 'Added'
    else prefix = 'Removed'

    return (
      <span>
        {prefix} from <Time time={n.start} zone={zone} /> to{' '}
        <Time time={n.end} zone={zone} omitSameDate={n.start} />
      </span>
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
    <OverrideDialogContext.Provider
      value={{
        onNewTempSched,
        setOverrideDialog,
      }}
    >
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
          <React.Fragment>
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
            {!isMobile && (
              <Button
                variant='contained'
                startIcon={<GroupAdd />}
                onClick={() =>
                  setOverrideDialog({
                    variantOptions: ['replace', 'remove', 'add', 'temp'],
                    removeUserReadOnly: false,
                  })
                }
                sx={{ ml: 1 }}
              >
                Create Override
              </Button>
            )}
          </React.Fragment>
        }
      />
      {isMobile && (
        <CreateFAB
          title='Create Override'
          onClick={() =>
            setOverrideDialog({
              variantOptions: ['replace', 'remove', 'add', 'temp'],
              removeUserReadOnly: false,
            })
          }
        />
      )}

      {/* create dialogs */}
      {overrideDialog && (
        <ScheduleOverrideDialog
          defaultValue={overrideDialog.defaultValue}
          variantOptions={overrideDialog.variantOptions}
          scheduleID={scheduleID}
          onClose={() => setOverrideDialog(null)}
          removeUserReadOnly={overrideDialog.removeUserReadOnly}
        />
      )}
      {configTempSchedule && (
        <TempSchedDialog
          value={configTempSchedule}
          onClose={() => setConfigTempSchedule(null)}
          scheduleID={scheduleID}
        />
      )}

      {/* edit dialogs by ID */}
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
    </OverrideDialogContext.Provider>
  )
}
