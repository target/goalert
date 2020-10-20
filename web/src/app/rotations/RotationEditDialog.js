import { gql } from '@apollo/client'
import React, { useState } from 'react'
import p from 'prop-types'
import { useQuery, useMutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import RotationForm from './RotationForm'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const query = gql`
  query($id: ID!) {
    rotation(id: $id) {
      id
      name
      description
      timeZone
      type
      shiftLength
      start
    }
  }
`

const mutation = gql`
  mutation($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

export default function RotationEditDialog(props) {
  const [value, setValue] = useState(null)

  const { loading, error, data } = useQuery(query, {
    variables: { id: props.rotationID },
    pollInterval: 0,
  })

  const [editRotation, editRotationStatus] = useMutation(mutation, {
    onCompleted: props.onClose,
  })

  if (loading && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  return (
    <FormDialog
      title='Edit Rotation'
      errors={nonFieldErrors(editRotationStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        editRotation({
          variables: {
            input: {
              id: props.rotationID,
              ...value,
            },
          },
        })
      }
      form={
        <RotationForm
          errors={fieldErrors(editRotationStatus.error)}
          disabled={editRotationStatus.loading}
          value={
            value || {
              name: data.rotation.name,
              description: data.rotation.description,
              timeZone: data.rotation.timeZone,
              type: data.rotation.type,
              shiftLength: data.rotation.shiftLength,
              start: data.rotation.start,
            }
          }
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}

RotationEditDialog.propTypes = {
  rotationID: p.string.isRequired,
  onClose: p.func,
}
