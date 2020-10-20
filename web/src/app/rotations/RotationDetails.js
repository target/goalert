import { gql, useQuery } from '@apollo/client'
import React, { useState } from 'react'
import p from 'prop-types'
import _ from 'lodash-es'

import PageActions from '../util/PageActions'
import OtherActions from '../util/OtherActions'
import CreateFAB from '../lists/CreateFAB'
import { handoffSummary } from './util'
import DetailsPage from '../details/DetailsPage'
import RotationEditDialog from './RotationEditDialog'
import RotationDeleteDialog from './RotationDeleteDialog'
import RotationUserList from './RotationUserList'
import RotationAddUserDialog from './RotationAddUserDialog'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'
import { Redirect } from 'react-router-dom'

import Spinner from '../loading/components/Spinner'
import { ObjectNotFound, GenericError } from '../error-pages'

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

  if (loading && !data) return <Spinner />
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
      <PageActions>
        <QuerySetFavoriteButton rotationID={rotationID} />
        <OtherActions
          actions={[
            { label: 'Edit Rotation', onClick: () => setShowEdit(true) },
            { label: 'Delete Rotation', onClick: () => setShowDelete(true) },
          ]}
        />
      </PageActions>
      <DetailsPage
        title={data.name}
        details={data.description}
        titleFooter={handoffSummary(data)}
        pageFooter={<RotationUserList rotationID={rotationID} />}
      />
    </React.Fragment>
  )
}

RotationDetails.propTypes = {
  rotationID: p.string.isRequired,
}
