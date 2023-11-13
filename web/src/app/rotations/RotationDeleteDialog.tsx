import React from 'react'
import { gql, useQuery, useMutation } from 'urql'
import Spinner from '../loading/components/Spinner'
import FormDialog from '../dialogs/FormDialog'
import { useLocation } from 'wouter'

import { get } from 'lodash'

const query = gql`
  query ($id: ID!) {
    rotation(id: $id) {
      id
      name
      description
      start
      timeZone
      type
    }
  }
`
const mutation = gql`
  mutation ($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function RotationDeleteDialog(props: {
  rotationID: string
  onClose: () => void
}): React.ReactNode {
  const [, navigate] = useLocation()
  const [{ data, fetching: dataLoading }] = useQuery({
    query,
    variables: { id: props.rotationID },
  })

  const [deleteRotationStatus, deleteRotation] = useMutation(mutation)

  if (!data && dataLoading) return <Spinner />

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the rotation: ${get(data, 'rotation.name')}`}
      loading={deleteRotationStatus.fetching}
      errors={deleteRotationStatus.error ? [deleteRotationStatus.error] : []}
      onClose={props.onClose}
      onSubmit={() =>
        deleteRotation(
          {
            input: [
              {
                id: props.rotationID,
                type: 'rotation',
              },
            ],
          },
          { additionalTypenames: ['Rotation'] },
        ).then((res) => {
          if (!res.error) {
            navigate('/rotations')
          }
        })
      }
    />
  )
}
