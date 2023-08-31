import React, { useState, useEffect } from 'react'
import { gql, useQuery, useMutation } from 'urql'

import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import RotationForm from './RotationForm'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { CreateRotationInput } from '../../schema'

const query = gql`
  query ($id: ID!) {
    rotation(id: $id) {
      id
      name
      description
      timeZone
      type
      shiftLength
      start
      nextHandoffTimes(num: 1)
    }
  }
`

const mutation = gql`
  mutation ($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

export default function RotationEditDialog(props: {
  rotationID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<CreateRotationInput | null>(null)

  const [{ fetching, error, data }] = useQuery({
    query,
    variables: { id: props.rotationID },
  })

  const [editRotationStatus, editRotation] = useMutation(mutation)

  useEffect(() => {
    if (!editRotationStatus.data) return
    props.onClose()
  }, [editRotationStatus.data])

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  return (
    <FormDialog
      title='Edit Rotation'
      errors={nonFieldErrors(editRotationStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        editRotation(
          {
            input: {
              id: props.rotationID,
              ...value,
            },
          },
          { additionalTypenames: ['Rotation'] },
        )
      }
      form={
        <RotationForm
          errors={fieldErrors(editRotationStatus.error)}
          value={
            value || {
              name: data.rotation.name,
              description: data.rotation.description,
              timeZone: data.rotation.timeZone,
              type: data.rotation.type,
              shiftLength: data.rotation.shiftLength,
              start: data.rotation.nextHandoffTimes[0] || data.rotation.start,
            }
          }
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
