import React from 'react'
import { gql, useQuery, useMutation } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const query = gql`
  query ($id: ID!) {
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
  mutation ($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`
const RotationUserDeleteDialog = (props: {
  rotationID: string
  userIndex: number
  onClose: () => void
}): JSX.Element => {
  const { rotationID, userIndex, onClose } = props
  const [, deleteUserMutation] = useMutation(mutation)
  const [{ fetching, data, error }] = useQuery({
    query: query,
    variables: {
      id: rotationID,
    },
  })

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const { userIDs, users, activeUserIndex } = data.rotation

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete ${
        users[userIndex] ? users[userIndex].name : null
      } from this rotation.`}
      onClose={onClose}
      onSubmit={() =>
        deleteUserMutation(
          {
            input: {
              id: rotationID,
              activeUserIndex:
                activeUserIndex > userIndex
                  ? activeUserIndex - 1
                  : activeUserIndex,
              userIDs: userIDs.filter(
                (_: string, index: number) => index !== userIndex,
              ),
            },
          },
          { additionalTypenames: ['Rotation'] },
        )
      }
    />
  )
}

export default RotationUserDeleteDialog
