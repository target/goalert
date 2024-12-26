import React, { Suspense, useState } from 'react'
import { useQuery, gql } from 'urql'
import FlatList, { FlatListListItem } from '../lists/FlatList'
import { Button, ButtonGroup, Card } from '@mui/material'
import { GroupAdd, PersonAdd } from '@mui/icons-material'
import Tooltip from '@mui/material/Tooltip'
import { startCase, sortBy } from 'lodash'
import { RotationAvatar, UserAvatar } from '../util/avatars'
import OtherActions from '../util/OtherActions'
import SpeedDial from '../util/SpeedDial'
import { AccountPlus, AccountMultiplePlus } from 'mdi-material-ui'
import ScheduleRuleCreateDialog from './ScheduleRuleCreateDialog'
import { ruleSummary } from './util'
import ScheduleRuleEditDialog from './ScheduleRuleEditDialog'
import ScheduleRuleDeleteDialog from './ScheduleRuleDeleteDialog'
import { GenericError } from '../error-pages'
import { DateTime } from 'luxon'
import { useScheduleTZ } from './useScheduleTZ'
import { useIsWidthDown } from '../util/useWidth'
import { ScheduleRule, ScheduleTarget, TargetType } from '../../schema'

const query = gql`
  query scheduleRules($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
      targets {
        target {
          id
          type
          name
        }
        rules {
          id
          start
          end
          weekdayFilter
        }
      }
    }
  }
`

interface ScheduleRuleListProps {
  scheduleID: string
}

interface TargetValue {
  id: string
  type: TargetType
}

export default function ScheduleRuleList(
  props: ScheduleRuleListProps,
): React.JSX.Element {
  const { scheduleID } = props
  const [editTarget, setEditTarget] = useState<TargetValue | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<TargetValue | null>(null)
  const [createType, setCreateType] = useState<TargetType | null>(null)
  const isMobile = useIsWidthDown('md')

  const [{ data, error }] = useQuery({
    query,
    variables: { id: scheduleID },
  })

  const { isLocalZone } = useScheduleTZ(scheduleID)

  if (error) {
    return <GenericError error={error.message} />
  }

  function renderSubText(rules: ScheduleRule[], timeZone: string): React.JSX.Element {
    const tzSummary = ruleSummary(rules, timeZone, timeZone)
    const tzAbbr = DateTime.local({ zone: timeZone }).toFormat('ZZZZ')
    const localTzSummary = ruleSummary(rules, timeZone, 'local')
    const localTzAbbr = DateTime.local({ zone: 'local' }).toFormat('ZZZZ')

    if (tzSummary === 'Always' || tzSummary === 'Never') {
      return tzSummary
    }

    return isLocalZone ? (
      <span aria-label='subtext'>{`${tzSummary} ${tzAbbr}`}</span>
    ) : (
      <Tooltip
        title={localTzSummary + ` ${localTzAbbr}`}
        placement='bottom-start'
        PopperProps={{
          'aria-label': 'local-timezone-tooltip',
        }}
      >
        <span aria-label='subtext'>{`${tzSummary} ${tzAbbr}`}</span>
      </Tooltip>
    )
  }

  function renderList(
    targets: ScheduleTarget[],
    timeZone: string,
  ): React.JSX.Element {
    const items: FlatListListItem[] = []

    let lastType: TargetType
    sortBy(targets, ['target.type', 'target.name']).forEach((tgt) => {
      const { name, id, type } = tgt.target
      if (type !== lastType) {
        items.push({ subHeader: startCase(type + 's') })
        lastType = type
      }

      items.push({
        title: name,
        url: (type === 'rotation' ? '/rotations/' : '/users/') + id,
        subText: renderSubText(tgt.rules, timeZone),
        icon:
          type === 'rotation' ? <RotationAvatar /> : <UserAvatar userID={id} />,
        secondaryAction: (
          <OtherActions
            actions={[
              {
                label: 'Edit',
                onClick: () => setEditTarget({ type, id }),
              },
              {
                label: 'Delete',
                onClick: () => setDeleteTarget({ type, id }),
              },
            ]}
          />
        ),
      })
    })

    return (
      <React.Fragment>
        <Card style={{ width: '100%', marginBottom: 64 }}>
          <FlatList
            headerNote={`Showing times in ${data.schedule.timeZone}.`}
            items={items}
            headerAction={
              !isMobile ? (
                <ButtonGroup variant='contained'>
                  <Button
                    startIcon={<GroupAdd />}
                    onClick={() => setCreateType('rotation')}
                  >
                    Add Rotation
                  </Button>
                  <Button
                    startIcon={<PersonAdd />}
                    onClick={() => setCreateType('user')}
                  >
                    Add User
                  </Button>
                </ButtonGroup>
              ) : undefined
            }
          />
        </Card>

        {isMobile && (
          <SpeedDial
            label='Add Assignment'
            actions={[
              {
                label: 'Add Rotation',
                onClick: () => setCreateType('rotation'),
                icon: <AccountMultiplePlus />,
              },
              {
                label: 'Add User',
                onClick: () => setCreateType('user'),
                icon: <AccountPlus />,
              },
            ]}
          />
        )}

        <Suspense>
          {createType && (
            <ScheduleRuleCreateDialog
              targetType={createType}
              scheduleID={scheduleID}
              onClose={() => setCreateType(null)}
            />
          )}
          {editTarget && (
            <ScheduleRuleEditDialog
              target={editTarget}
              scheduleID={scheduleID}
              onClose={() => setEditTarget(null)}
            />
          )}
          {deleteTarget && (
            <ScheduleRuleDeleteDialog
              target={deleteTarget}
              scheduleID={scheduleID}
              onClose={() => setDeleteTarget(null)}
            />
          )}
        </Suspense>
      </React.Fragment>
    )
  }

  return renderList(data.schedule.targets, data.schedule.timeZone)
}
