import React from 'react'
import { useDispatch } from 'react-redux'
import p from 'prop-types'
import gql from 'graphql-tag'
import Spinner from '../loading/components/Spinner'
import FormDialog from '../dialogs/FormDialog'
import { useQuery, useMutation } from 'react-apollo'
import { push } from 'connected-react-router'

const query = gql`
  query($id: ID!) {
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
  mutation($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function RotationDeleteDialog(props) {
  const dispatch = useDispatch()
  const { data, loading: dataLoading } = useQuery(query, {
    variables: { id: props.rotationID },
  })
  const [deleteRotation, deleteRotationStatus] = useMutation(mutation, {
    variables: {
      input: [
        {
          id: props.rotationID,
          type: 'rotation',
        },
      ],
    },
    onCompleted: () => dispatch(push('/rotations')),
  })

  if (dataLoading) return <Spinner />

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the rotation: ${data.rotation.name}`}
      loading={deleteRotationStatus.loading}
      errors={deleteRotationStatus.error ? [deleteRotationStatus.error] : []}
      onClose={props.onClose}
      onSubmit={() => deleteRotation()}
    />
  )
}

RotationDeleteDialog.propTypes = {
  rotationID: p.string.isRequired,
  onClose: p.func,
}
