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
const RotationSetActiveDialog = (props) => {
  const { rotationID, userIndex, onClose } = props
  const { loading, data, error } = useQuery(query, {
    pollInterval: 0,
    variables: {
      id: rotationID,
    },
  })
  const [setActiveMutation] = useMutation(mutation, {
    onCompleted: onClose,
    variables: {
      input: {
        id: rotationID,
        activeUserIndex: userIndex,
      },
    },
  })

  if (loading && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />
  const { users } = data.rotation

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will set ${users[userIndex].name} active on this rotation.`}
      onClose={onClose}
      onSubmit={() => setActiveMutation()}
    />
  )
}

RotationSetActiveDialog.propTypes = {
  rotationID: p.string.isRequired,
  userIndex: p.number.isRequired,
  onClose: p.func.isRequired,
}

export default RotationSetActiveDialog
