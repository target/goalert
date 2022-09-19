import React, { useState } from 'react'
import { gql, useQuery } from 'urql'
import _ from 'lodash'
import { Redirect } from 'wouter'
import { Edit, Delete } from '@mui/icons-material'

import CreateFAB from '../lists/CreateFAB'
import { handoffSummary } from './util'
import DetailsPage from '../details/DetailsPage'
import RotationEditDialog from './RotationEditDialog'
import RotationDeleteDialog from './RotationDeleteDialog'
import RotationUserList from './RotationUserList'
import RotationAddUserDialog from './RotationAddUserDialog'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'
import Spinner from '../loading/components/Spinner'
import { ObjectNotFound, GenericError } from '../error-pages'
import { RotationAvatar } from '../util/avatars'

const query = gql`
  fragment RotationTitleQuery on Rotation {
    id
    name
    description
  }

  query rotationDetails($id: ID!) {
    rotation(id: $id) {
      ...RotationTitleQuery

      activeUserIndex
      userIDs
      type
      shiftLength
      timeZone
      start
    }
  }
`

export default function RotationDetails(props: {
  rotationID: string
}): JSX.Element {
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [showAddUser, setShowAddUser] = useState(false)

  const [{ data: _data, fetching, error }] = useQuery({
    query,
    variables: { id: props.rotationID },
  })

  const data = _.get(_data, 'rotation', null)

  if (fetching && !data?.name) return <Spinner />
  if (error) return <GenericError error={error.message} />

  if (!data)
    return showDelete ? (
      <Redirect to='/rotations' />
    ) : (
      <ObjectNotFound type='rotation' />
    )

  return (
    <React.Fragment>
      <CreateFAB title='Add User' onClick={() => setShowAddUser(true)} />
      {showAddUser && (
        <RotationAddUserDialog
          rotationID={props.rotationID}
          userIDs={data.userIDs}
          onClose={() => setShowAddUser(false)}
        />
      )}
      {showEdit && (
        <RotationEditDialog
          rotationID={props.rotationID}
          onClose={() => setShowEdit(false)}
        />
      )}
      {showDelete && (
        <RotationDeleteDialog
          rotationID={props.rotationID}
          onClose={() => setShowDelete(false)}
        />
      )}
      <DetailsPage
        avatar={<RotationAvatar />}
        title={data.name}
        subheader={handoffSummary(data)}
        details={data.description}
        pageContent={<RotationUserList rotationID={props.rotationID} />}
        secondaryActions={[
          {
            label: 'Edit',
            icon: <Edit />,
            handleOnClick: () => setShowEdit(true),
          },
          {
            label: 'Delete',
            icon: <Delete />,
            handleOnClick: () => setShowDelete(true),
          },
          <QuerySetFavoriteButton
            key='secondary-action-favorite'
            id={props.rotationID}
            type='rotation'
          />,
        ]}
      />
    </React.Fragment>
  )
}
