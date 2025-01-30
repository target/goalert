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
      activeUserIndex
      users {
        id
        name
      }
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
  const [deleteUserMutationStatus, deleteUserMutation] = useMutation(mutation)
  const [{ fetching, data, error }] = useQuery({
    query,
    variables: {
      id: rotationID,
    },
  })

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const { userIDs, users } = data.rotation

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete ${
        users[userIndex] ? users[userIndex].name : null
      } from this rotation.`}
      onClose={onClose}
      errors={
        deleteUserMutationStatus.error ? [deleteUserMutationStatus.error] : []
      }
      onSubmit={() =>
        deleteUserMutation(
          {
            input: {
              id: rotationID,
              userIDs: userIDs.filter(
                (_: string, index: number) => index !== userIndex,
              ),
              activeUserIndex:
                userIndex < data.rotation.activeUserIndex
                  ? data.rotation.activeUserIndex - 1
                  : data.rotation.activeUserIndex,
            },
          },
          { additionalTypenames: ['Rotation'] },
        ).then((res) => {
          if (res.error) return
          onClose()
        })
      }
    />
  )
}

export default RotationUserDeleteDialog
