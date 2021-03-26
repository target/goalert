import React from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import Query from '../util/Query'
import { Mutation } from '@apollo/client/react/components'
import FormDialog from '../dialogs/FormDialog'

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
  const renderDialog = (data, commit) => {
    const { userIDs, users } = data
    const { rotationID, userIndex, onClose } = props

    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete ${
          users[userIndex] ? users[userIndex].name : null
        } from this rotation.`}
        onClose={onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
                id: rotationID,
                userIDs: userIDs.filter((id, index) => index !== userIndex),
              },
            },
          })
        }}
      />
    )
  }

  const renderMutation = (data) => {
    return (
      <Mutation mutation={mutation} onCompleted={props.onClose}>
        {(commit) => renderDialog(data, commit)}
      </Mutation>
    )
  }

  return (
    <Query
      query={query}
      variables={{ id: props.rotationID }}
      render={({ data }) => renderMutation(data.rotation)}
    />
  )
}

RotationUserDeleteDialog.propTypes = {
  rotationID: p.string.isRequired,
  userIndex: p.number.isRequired,
  onClose: p.func.isRequired,
}

export default RotationUserDeleteDialog
