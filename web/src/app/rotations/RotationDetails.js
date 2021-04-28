import React, { useState } from 'react'
import { gql, useQuery } from '@apollo/client'
import p from 'prop-types'
import _ from 'lodash'
import { Redirect } from 'react-router-dom'
import { Edit, Delete } from '@material-ui/icons'

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

export default function RotationDetails({ rotationID }) {
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [showAddUser, setShowAddUser] = useState(false)

  const { data: _data, loading, error } = useQuery(query, {
    variables: { id: rotationID },
    returnPartialData: true,
  })

  const data = _.get(_data, 'rotation', null)

  if (loading && !data?.name) return <Spinner />
  if (error) return <GenericError error={error.message} />

  if (!data)
    return showDelete ? <Redirect to='/rotations' push /> : <ObjectNotFound />

  return (
    <React.Fragment>
      <CreateFAB title='Add User' onClick={() => setShowAddUser(true)} />
      {showAddUser && (
        <RotationAddUserDialog
          rotationID={rotationID}
          userIDs={data.userIDs}
          onClose={() => setShowAddUser(false)}
        />
      )}
      {showEdit && (
        <RotationEditDialog
          rotationID={rotationID}
          onClose={() => setShowEdit(false)}
        />
      )}
      {showDelete && (
        <RotationDeleteDialog
          rotationID={rotationID}
          onClose={() => setShowDelete(false)}
        />
      )}
      <DetailsPage
        avatar={<RotationAvatar />}
        title={data.name}
        subheader={handoffSummary(data)}
        details={data.description}
        pageContent={<RotationUserList rotationID={rotationID} />}
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
            rotationID={rotationID}
          />,
        ]}
      />
    </React.Fragment>
  )
}

RotationDetails.propTypes = {
  rotationID: p.string.isRequired,
}
