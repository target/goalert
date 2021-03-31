import React, { ReactNode, useState } from 'react'
import { gql, useQuery } from '@apollo/client'
import FlatList from '../lists/FlatList'
import { ScheduleTZFilter } from './ScheduleTZFilter'
import { Grid, Card } from '@material-ui/core'
import FilterContainer from '../util/FilterContainer'
import PageActions from '../util/PageActions'
import { startCase, sortBy } from 'lodash'
import { RotationAvatar, UserAvatar } from '../util/avatars'
import OtherActions from '../util/OtherActions'
import SpeedDial from '../util/SpeedDial'
import { AccountPlus, AccountMultiplePlus } from 'mdi-material-ui'
import ScheduleRuleCreateDialog from './ScheduleRuleCreateDialog'
import { ruleSummary } from './util'
import ScheduleRuleEditDialog from './ScheduleRuleEditDialog'
import ScheduleRuleDeleteDialog from './ScheduleRuleDeleteDialog'
import { useResetURLParams, useURLParam } from '../actions'
import { GenericError } from '../error-pages/Errors'
import Spinner from '../loading/components/Spinner'
import { Query, Target, TargetType } from '../../schema'

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

interface FlatListItem {
  highlight?: boolean
  title: ReactNode
  subText?: ReactNode
  secondaryAction?: ReactNode
  url?: string
  icon?: ReactNode
  id?: string
}

interface FlatListSubHeader {
  subHeader: ReactNode
}

function ScheduleRuleList(props: ScheduleRuleListProps): JSX.Element {
  const { scheduleID } = props
  const [zone] = useURLParam('tz', 'local')
  const resetFilter = useResetURLParams('tz')
  const [editTarget, setEditTarget] = useState<Target | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Target | null>(null)
  const [createType, setCreateType] = useState<string | null>(null)

  const { data, loading, error } = useQuery<Query>(query, {
    variables: { id: scheduleID },
  })

  if (error) {
    return <GenericError error={error.message} />
  }

  if (loading && !data) {
    return <Spinner />
  }

  const targets = data?.schedule?.targets
  const timeZone = data?.schedule?.timeZone
  const items: (FlatListItem | FlatListSubHeader)[] = []

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
      subText: ruleSummary(tgt.rules, timeZone, zone),
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
      <PageActions>
        <FilterContainer onReset={() => resetFilter()}>
          <Grid item xs={12}>
            <ScheduleTZFilter scheduleID={scheduleID} />
          </Grid>
        </FilterContainer>
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
      </PageActions>
      <Card style={{ width: '100%', marginBottom: 64 }}>
        <FlatList
          headerNote={`Showing times in ${
            zone === 'local' ? 'local time' : zone
          }.`}
          items={items}
        />
      </Card>

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
    </React.Fragment>
  )
}

export default ScheduleRuleList
