import React, { useEffect, useState } from 'react'
import p from 'prop-types'
import { gql, useMutation, useQuery } from '@apollo/client'
import { DateTime } from 'luxon'
import Card from '@mui/material/Card'
import makeStyles from '@mui/styles/makeStyles'
import CardHeader from '@mui/material/CardHeader'

import FlatList from '../lists/FlatList'
import { reorderList, calcNewActiveIndex } from './util'
import OtherActions from '../util/OtherActions'
import CountDown from '../util/CountDown'
import RotationSetActiveDialog from './RotationSetActiveDialog'
import RotationUserDeleteDialog from './RotationUserDeleteDialog'
import { UserAvatar } from '../util/avatars'
import { styles as globalStyles } from '../styles/materialStyles'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'

const query = gql`
  query rotationUsers($id: ID!) {
    rotation(id: $id) {
      id
      users {
        id
        name
      }
      activeUserIndex
      nextHandoffTimes
    }
  }
`

const mutation = gql`
  mutation updateRotation($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

const useStyles = makeStyles((theme) => {
  const { cardHeader } = globalStyles(theme)

  return {
    cardHeader,
  }
})

function RotationUserList({ rotationID }) {
  const classes = useStyles()
  const [deleteIndex, setDeleteIndex] = useState(null)
  const [setActiveIndex, setSetActiveIndex] = useState(null)
  const [lastSwap, setLastSwap] = useState([])

  const {
    data,
    loading: qLoading,
    error: qError,
  } = useQuery(query, {
    variables: { id: rotationID },
  })

  const [updateRotation, { error: mError }] = useMutation(mutation)

  // reset swap history on add/remove participant
  useEffect(() => {
    setLastSwap([])
  }, [data?.rotation?.users?.length])

  if (qLoading && !data) return <Spinner />
  if (data && !data.rotation) return <ObjectNotFound type='rotation' />
  if (qError || mError)
    return <GenericError error={qError.message || mError.message} />

  const { users, activeUserIndex, nextHandoffTimes } = data.rotation

  // duplicate first entry
  const _nextHandoffTimes = (nextHandoffTimes || [])
    .slice(0, 1)
    .concat(nextHandoffTimes)

  const handoff = users.map((u, index) => {
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
    }
    return (
      'Starts at ' +
      DateTime.fromISO(time).toLocaleString(DateTime.TIME_SIMPLE) +
      ' ' +
      DateTime.fromISO(time).toRelativeCalendar()
    )
  })

  // re-enact swap history to get unique identier per list item
  let listIDs = users.map((_, idx) => idx)
  lastSwap.forEach((s) => {
    listIDs = reorderList(listIDs, s.oldIndex, s.newIndex)
  })

  return (
    <React.Fragment>
      {deleteIndex !== null && (
        <RotationUserDeleteDialog
          rotationID={rotationID}
          userIndex={deleteIndex}
          onClose={() => setDeleteIndex(null)}
        />
      )}
      {setActiveIndex !== null && (
        <RotationSetActiveDialog
          rotationID={rotationID}
          userIndex={setActiveIndex}
          onClose={() => setSetActiveIndex(null)}
        />
      )}
      <Card>
        <CardHeader
          className={classes.cardHeader}
          component='h3'
          title='Users'
        />

        <FlatList
          data-cy='users'
          emptyMessage='No users currently assigned to this rotation'
          headerNote={
            users.length ? "Click and drag on a user's name to re-order" : ''
          }
          items={users.map((u, index) => ({
            title: u.name,
            id: String(listIDs[index]),
            highlight: index === activeUserIndex,
            icon: <UserAvatar userID={u.id} />,
            subText: handoff[index],
            secondaryAction: (
              <OtherActions
                actions={[
                  {
                    label: 'Set Active',
                    onClick: () => setSetActiveIndex(index),
                  },
                  {
                    label: 'Remove',
                    onClick: () => setDeleteIndex(index),
                  },
                ]}
              />
            ),
          }))}
          onReorder={(oldIndex, newIndex) => {
            setLastSwap(lastSwap.concat({ oldIndex, newIndex }))

            const updatedUsers = reorderList(
              users.map((u) => u.id),
              oldIndex,
              newIndex,
            )
            const newActiveIndex = calcNewActiveIndex(
              activeUserIndex,
              oldIndex,
              newIndex,
            )
            const params = { id: rotationID, userIDs: updatedUsers }

            if (newActiveIndex !== -1) {
              params.activeUserIndex = newActiveIndex
            }

            return updateRotation({
              variables: { input: params },
              update: (cache, response) => {
                if (!response.data.updateRotation) {
                  return
                }
                const data = cache.readQuery({
                  query,
                  variables: { id: rotationID },
                })

                const users = reorderList(
                  data.rotation.users,
                  oldIndex,
                  newIndex,
                )

                cache.writeQuery({
                  query,
                  variables: { id: rotationID },
                  data: {
                    ...data,
                    rotation: {
                      ...data.rotation,
                      activeUserIndex:
                        newActiveIndex === -1
                          ? data.rotation.activeUserIndex
                          : newActiveIndex,
                      users,
                    },
                  },
                })
              },
              optimisticResponse: {
                __typename: 'Mutation',
                updateRotation: true,
              },
            })
          }}
        />
      </Card>
    </React.Fragment>
  )
}

RotationUserList.propTypes = {
  rotationID: p.string.isRequired,
}

export default RotationUserList
