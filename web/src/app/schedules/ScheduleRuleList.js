import React, { useState } from 'react'
import p from 'prop-types'
import { gql, useQuery } from '@apollo/client'
import FlatList from '../lists/FlatList'
import { ScheduleTZFilter } from './ScheduleTZFilter'
import { Grid, Card } from '@material-ui/core'
import FilterContainer from '../util/FilterContainer'
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
import { GenericError } from '../error-pages'
import Spinner from '../loading/components/Spinner'

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

export default function ScheduleRuleList(props) {
  const [editTarget, setEditTarget] = useState(null)
  const [deleteTarget, setDeleteTarget] = useState(null)
  const [createType, setCreateType] = useState(null)

  const [zone] = useURLParam('tz', 'local')
  const resetFilter = useResetURLParams('tz')

  const { data, loading, error } = useQuery(query, {
    variables: { id: props.scheduleID },
    pollInterval: 0,
  })

  if (error) {
    return <GenericError error={error.message} />
  }

  if (loading && !data) {
    return <Spinner />
  }

  function getHeaderNote() {
    return `Showing times in ${zone === 'local' ? 'local time' : zone}.`
  }

  function renderList(targets, timeZone) {
    const items = []

    let lastType
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
        <Card style={{ width: '100%', marginBottom: 64 }}>
          <FlatList
            headerNote={getHeaderNote()}
            headerAction={
              <FilterContainer onReset={() => resetFilter()}>
                <Grid item xs={12}>
                  <ScheduleTZFilter scheduleID={props.scheduleID} />
                </Grid>
              </FilterContainer>
            }
            items={items}
          />
        </Card>

        {createType && (
          <ScheduleRuleCreateDialog
            targetType={createType}
            scheduleID={props.scheduleID}
            onClose={() => setCreateType(null)}
          />
        )}
        {editTarget && (
          <ScheduleRuleEditDialog
            target={editTarget}
            scheduleID={props.scheduleID}
            onClose={() => setEditTarget(null)}
          />
        )}
        {deleteTarget && (
          <ScheduleRuleDeleteDialog
            target={deleteTarget}
            scheduleID={props.scheduleID}
            onClose={() => setDeleteTarget(null)}
          />
        )}
      </React.Fragment>
    )
  }

  return renderList(data.schedule.targets, data.schedule.timeZone)
}

ScheduleRuleList.propTypes = {
  scheduleID: p.string.isRequired,
}
