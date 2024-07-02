import React, { Suspense, useCallback, useState } from 'react'
import { Button, Grid, FormControlLabel, Switch, Tooltip } from '@mui/material'
import { GroupAdd } from '@mui/icons-material'
import { DateTime } from 'luxon'
import { gql, useQuery } from 'urql'
import { UserAvatar } from '../util/avatars'
import OtherActions from '../util/OtherActions'
import FilterContainer from '../util/FilterContainer'
import { UserSelect } from '../selection'
import { useURLParam, useResetURLParams } from '../actions'
import ScheduleOverrideDeleteDialog from './ScheduleOverrideDeleteDialog'
import { formatOverrideTime } from './util'
import ScheduleOverrideEditDialog from './ScheduleOverrideEditDialog'
import { useScheduleTZ } from './useScheduleTZ'
import { useIsWidthDown } from '../util/useWidth'
import { OverrideDialogContext } from './ScheduleDetails'
import TempSchedDialog from './temp-sched/TempSchedDialog'
import { defaultTempSchedValue } from './temp-sched/sharedUtils'
import ScheduleOverrideDialog from './ScheduleOverrideDialog'
import CreateFAB from '../lists/CreateFAB'
import ListPageControls from '../lists/ListPageControls'
import FlatList from '../lists/FlatList'

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
const context = { suspense: false }

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

  const { zone, isLocalZone } = useScheduleTZ(scheduleID)
  const onNewTempSched = useCallback(
    () => setConfigTempSchedule(defaultTempSchedValue(zone)),
    [],
  )
  const [cursor, setCursor] = useState('')

  const inputVars = {
    scheduleID,
    start: showPast ? null : now,
    filterAnyUserID: userFilter,
    after: cursor,
  }

  const [q] = useQuery({
    query,
    variables: { input: inputVars },
    context,
  })
  const nextCursor = q.data?.userOverrides.pageInfo.hasNextPage
    ? q.data?.userOverrides.pageInfo.endCursor
    : ''

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
    <OverrideDialogContext.Provider
      value={{
        onNewTempSched,
        setOverrideDialog,
      }}
    >
      <ListPageControls
        nextCursor={nextCursor}
        onCursorChange={setCursor}
        loading={q.fetching}
        slots={{
          list: (
            <FlatList
              emptyMessage='No results'
              headerNote={note}
              items={
                q.data?.userOverrides.nodes.map((n) => ({
                  title: n.addUser ? n.addUser.name : n.removeUser.name,
                  subText: subText(n),
                  icon: (
                    <UserAvatar
                      userID={n.addUser ? n.addUser.id : n.removeUser.id}
                    />
                  ),
                  secondaryAction: (
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
                })) || []
              }
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
          ),
        }}
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
      <Suspense>
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
      </Suspense>
    </OverrideDialogContext.Provider>
  )
}
