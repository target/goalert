import React, { useState } from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import FlatList from '../lists/FlatList'
import Card from '@material-ui/core/Card'
import CardHeader from '@material-ui/core/CardHeader'
import OtherActions from '../util/OtherActions'
import CountDown from '../util/CountDown'
import RotationSetActiveDialog from './RotationSetActiveDialog'
import RotationUserDeleteDialog from './RotationUserDeleteDialog'
import { DateTime } from 'luxon'
import { UserAvatar } from '../util/avatar'
import { makeStyles } from '@material-ui/core'
import { styles as globalStyles } from '../styles/materialStyles'
import { useMutation } from 'react-apollo'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'
import Dialog from '@material-ui/core/Dialog'
import { query as rotationDetailsQuery } from './RotationDetails'

const updateRotMutation = gql`
  mutation updateRotation($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

const useStyles = makeStyles(theme => ({
  cardHeader: globalStyles(theme).cardHeader,
}))

export default function RotationUsersList(props) {
  const { users, activeUserIndex, nextHandoffTimes } = props.rotation

  let oldID = null
  let oldIdx = null
  let newIdx = null
  let nextActiveIdx = null
  let userIDs = users.map(user => user.id)

  const classes = useStyles()
  const [showErrorDialog, setShowErrorDialog] = useState(false)
  const [deleteIndex, setDeleteIndex] = useState(null)
  const [activeIndex, setActiveIndex] = useState(null)
  const [updateRot, updateRotStatus] = useMutation(updateRotMutation, {
    onCompleted: () => {
      oldID = null
      oldIdx = null
      newIdx = null
      nextActiveIdx = null
    },
    onError: () => setShowErrorDialog(true),
    optimisticResponse: {
      updateRotation: true,
    },
    update: (cache, { data }) => updateCache(cache, data),
  })

  function arrayMove(arr) {
    const el = arr[oldIdx]
    arr.splice(oldIdx, 1)
    arr.splice(newIdx, 0, el)
  }

  /*
   * Executes on drag end. Once the mutation completes
   * successfully, updateCache will be called to update
   * the UI with the correct data.
   */
  function onReorder(result) {
    // dropped outside the list
    if (!result.destination) {
      return
    }

    oldID = result.draggableId
    oldIdx = userIDs.indexOf(oldID)
    newIdx = result.destination.index

    // if moving the active user, keep them as active
    // otherwise, ignore
    nextActiveIdx = activeUserIndex === oldIdx ? newIdx : activeUserIndex

    // re-order sids array
    arrayMove(userIDs)

    // call mutation
    return updateRot({
      variables: {
        input: {
          id: props.rotationID,
          userIDs,
          activeUserIndex: nextActiveIdx,
        },
      },
    })
  }

  function updateCache(cache, data) {
    // mutation returns true on a success
    if (!data.updateRotation || oldIdx == null || newIdx == null) {
      return
    }

    // variables for query to read/write from the cache
    const variables = {
      rotationID: props.rotationID,
    }

    // get the current state of the steps in the cache
    const { rotation } = cache.readQuery({
      query: rotationDetailsQuery,
      variables,
    })

    // get steps from cache
    const users = rotation.users.slice()

    // if optimistic cache update was successful, return out
    if (users[newIdx].id === oldID) return

    // re-order escalationPolicy.steps array
    arrayMove(users)

    // write new steps order to cache
    cache.writeQuery({
      query: rotationDetailsQuery,
      variables,
      data: {
        rotation: {
          ...rotation,
          users,
          activeUserIndex: nextActiveIdx,
        },
      },
    })
  }

  // main render return
  return (
    <React.Fragment>
      <Card>
        <CardHeader
          className={classes.cardHeader}
          component='h3'
          title='Users'
        />
        {renderUsersList()}
      </Card>
      {deleteIndex !== null && (
        <RotationUserDeleteDialog
          rotationID={props.rotationID}
          userIndex={deleteIndex}
          onClose={() => setDeleteIndex(null)}
        />
      )}
      {activeIndex !== null && (
        <RotationSetActiveDialog
          rotationID={props.rotationID}
          userIndex={activeIndex}
          onClose={() => setActiveIndex(null)}
        />
      )}
      <Dialog open={showErrorDialog} onClose={() => setShowErrorDialog(false)}>
        <DialogTitleWrapper title='An error occurred' />
        <DialogContentError
          error={updateRotStatus.error && updateRotStatus.error.message}
        />
      </Dialog>
    </React.Fragment>
  )

  function renderUsersList() {
    const handoffData = getHandoffData()

    return (
      <FlatList
        data-cy='users'
        emptyMessage='No users currently assigned to this rotation'
        headerNote={
          users.length ? "Click and drag on a user's name to re-order" : ''
        }
        onReorder={onReorder}
        items={users.map((u, index) => ({
          title: u.name,
          id: u.id,
          highlight: index === activeUserIndex,
          icon: <UserAvatar userID={u.id} />,
          subText: handoffData[index],
          secondaryAction: (
            <OtherActions
              actions={[
                {
                  label: 'Set Active',
                  onClick: () => setActiveIndex(index),
                },
                {
                  label: 'Remove',
                  onClick: () => setDeleteIndex(index),
                },
              ]}
            />
          ),
        }))}
      />
    )
  }

  function getHandoffData() {
    // duplicate first entry
    const _nextHandoffTimes = (nextHandoffTimes || [])
      .slice(0, 1)
      .concat(nextHandoffTimes)

    return users.map((u, index) => {
      const handoffIndex =
        (index + (users.length - activeUserIndex)) % users.length

      const time = _nextHandoffTimes[handoffIndex]

      if (!time) {
        return null
      }

      if (index === activeUserIndex) {
        return (
          <CountDown
            end={time}
            weeks
            days
            hours
            minutes
            prefix='Active for the next '
            style={{ marginLeft: '1em' }}
            expiredTimeout={60}
            expiredMessage='< 1 Minute'
          />
        )
      } else {
        return (
          'Starts at ' +
          DateTime.fromISO(time).toLocaleString(DateTime.TIME_SIMPLE) +
          ' ' +
          DateTime.fromISO(time).toRelativeCalendar()
        )
      }
    })
  }
}

RotationUsersList.propTypes = {
  rotationID: p.string.isRequired,
  rotation: p.shape({
    users: p.arrayOf(
      p.shape({
        id: p.string.isRequired,
        name: p.string.isRequired,
      }),
    ).isRequired,
    activeUserIndex: p.number.isRequired,
    nextHandoffTimes: p.array.isRequired,
  }).isRequired,
}
