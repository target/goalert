import React from 'react'
import { gql, useQuery, useMutation } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const query = gql`
  query ($id: ID!) {
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
  mutation ($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`
const RotationSetActiveDialog = (props: {
  rotationID: string
  userIndex: number
  onClose: () => void
}): React.JSX.Element => {
  const { rotationID, userIndex, onClose } = props
  const [{ fetching, data, error }] = useQuery({
    query,
    variables: {
      id: rotationID,
    },
  })
  const [, setActiveMutation] = useMutation(mutation)

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />
  const { users } = data.rotation

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will set ${users[userIndex].name} active on this rotation.`}
      onClose={onClose}
      onSubmit={() =>
        setActiveMutation(
          {
            input: {
              id: rotationID,
              activeUserIndex: userIndex,
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

export default RotationSetActiveDialog
