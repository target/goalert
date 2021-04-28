import React from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const query = gql`
  query($id: ID!) {
    rotation(id: $id) {
      id
      userIDs
      users {
        id
        name
      }
      activeUserIndex
    }
  }
`

const mutation = gql`
  mutation($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`
const RotationUserDeleteDialog = (props) => {
  const { rotationID, userIndex, onClose } = props
  const { loading, data, error } = useQuery(query, {
    pollInterval: 0,
    variables: {
      id: rotationID,
    },
  })
  const { userIDs, users } = data.rotation
  const [deleteUserMutation] = useMutation(mutation, {
    onCompleted: onClose,
    variables: {
      input: {
        id: rotationID,
        userIDs: userIDs.filter((_, index) => index !== userIndex),
      },
    },
  })

  if (loading && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete ${
        users[userIndex] ? users[userIndex].name : null
      } from this rotation.`}
      onClose={onClose}
      onSubmit={() => deleteUserMutation()}
    />
  )
}

RotationUserDeleteDialog.propTypes = {
  rotationID: p.string.isRequired,
  userIndex: p.number.isRequired,
  onClose: p.func.isRequired,
}

export default RotationUserDeleteDialog
