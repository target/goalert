import React, { useEffect, useState } from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import makeStyles from '@mui/styles/makeStyles'
import CardHeader from '@mui/material/CardHeader'
import { Theme } from '@mui/material/styles'
import { Add } from '@mui/icons-material'
import FlatList from '../lists/FlatList'
import { reorderList, calcNewActiveIndex } from './util'
import OtherActions from '../util/OtherActions'
import RotationSetActiveDialog from './RotationSetActiveDialog'
import RotationUserDeleteDialog from './RotationUserDeleteDialog'
import { UserAvatar } from '../util/avatars'
import { styles as globalStyles } from '../styles/materialStyles'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'
import { User, Rotation } from '../../schema'
import { Time } from '../util/Time'
import CreateFAB from '../lists/CreateFAB'
import RotationAddUserDialog from './RotationAddUserDialog'
import { useIsWidthDown } from '../util/useWidth'

const query = gql`
  query rotationUsers($id: ID!) {
    rotation(id: $id) {
      id
      users {
        id
        name
      }
      timeZone
      activeUserIndex
      nextHandoffTimes
      userIDs
    }
  }
`

const mutation = gql`
  mutation updateRotation($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

const useStyles = makeStyles((theme: Theme) => {
  const { cardHeader } = globalStyles(theme)

  return {
    cardHeader,
  }
})

interface RotationUserListProps {
  rotationID: string
}

type SwapType = {
  oldIndex: number
  newIndex: number
}

function RotationUserList(props: RotationUserListProps): JSX.Element {
  const classes = useStyles()
  const { rotationID } = props
  const [deleteIndex, setDeleteIndex] = useState<number | null>(null)
  const [setActiveIndex, setSetActiveIndex] = useState<number | null>(null)
  const [showAddUser, setShowAddUser] = useState(false)
  const [lastSwap, setLastSwap] = useState<SwapType[]>([])
  const isMobile = useIsWidthDown('md')

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
    return <GenericError error={qError?.message || mError?.message} />

  const { users, userIDs, activeUserIndex, nextHandoffTimes } = data.rotation

  // duplicate first entry
  const _nextHandoffTimes = (nextHandoffTimes || [])
    .slice(0, 1)
    .concat(nextHandoffTimes)

  const handoff = users.map((u: User, index: number) => {
    const handoffIndex =
      (index + (users.length - activeUserIndex)) % users.length
    const time = _nextHandoffTimes[handoffIndex]
    if (!time) {
      return null
    }

    if (index === activeUserIndex) {
      return (
        <Time
          key={index}
          prefix='Shift ends '
          time={time}
          zone={data?.rotation?.timeZone}
          format='relative'
          units={['years', 'months', 'weeks', 'days', 'hours', 'minutes']}
          precise
        />
      )
    }
    return (
      <Time
        key={index}
        prefix='Starts '
        time={time}
        zone={data?.rotation?.timeZone}
        format='relative'
        units={['years', 'months', 'weeks', 'days', 'hours', 'minutes']}
        precise
      />
    )
  })

  // re-enact swap history to get unique identier per list item
  let listIDs = users.map((_: User, idx: number) => idx)
  lastSwap.forEach((s) => {
    listIDs = reorderList(listIDs, s.oldIndex, s.newIndex)
  })

  return (
    <React.Fragment>
      {isMobile && (
        <CreateFAB title='Add User' onClick={() => setShowAddUser(true)} />
      )}
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
      {showAddUser && (
        <RotationAddUserDialog
          rotationID={rotationID}
          userIDs={userIDs ?? []}
          onClose={() => setShowAddUser(false)}
        />
      )}
      <Card>
        <CardHeader
          className={classes.cardHeader}
          component='h3'
          title='Users'
          action={
            !isMobile ? (
              <Button
                variant='contained'
                onClick={() => setShowAddUser(true)}
                startIcon={<Add />}
              >
                Add User
              </Button>
            ) : null
          }
        />

        <FlatList
          data-cy='users'
          emptyMessage='No users currently assigned to this rotation'
          headerNote={users.length ? 'Toggle edit to reorder users' : ''}
          toggleDnD
          items={users.map((u: User, index: number) => ({
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
          onReorder={(oldIndex: number, newIndex: number) => {
            setLastSwap(lastSwap.concat({ oldIndex, newIndex }))

            const updatedUsers = reorderList(
              users.map((u: User) => u.id),
              oldIndex,
              newIndex,
            )
            const newActiveIndex = calcNewActiveIndex(
              activeUserIndex,
              oldIndex,
              newIndex,
            )
            const params = {
              id: rotationID,
              userIDs: updatedUsers,
              activeUserIndex,
            }

            if (newActiveIndex !== -1) {
              params.activeUserIndex = newActiveIndex
            }

            return updateRotation({
              variables: { input: params },
              update: (cache, response) => {
                if (!response.data.updateRotation) {
                  return
                }
                const data: { rotation: Rotation } | null = cache.readQuery({
                  query,
                  variables: { id: rotationID },
                })

                if (data?.rotation?.users) {
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
                        ...data?.rotation,
                        activeUserIndex:
                          newActiveIndex === -1
                            ? data?.rotation?.activeUserIndex
                            : newActiveIndex,
                        users,
                      },
                    },
                  })
                }
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
export default RotationUserList
