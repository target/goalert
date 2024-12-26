import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import RotationForm from './RotationForm'
import { DateTime } from 'luxon'
import { Redirect } from 'wouter'
import { CreateRotationInput } from '../../schema'

const mutation = gql`
  mutation ($input: CreateRotationInput!) {
    createRotation(input: $input) {
      id
      name
      description
      start
      timeZone
      type
      shiftLength
    }
  }
`

const RotationCreateDialog = (props: { onClose?: () => void }): React.JSX.Element => {
  const [value, setValue] = useState<CreateRotationInput>({
    name: '',
    description: '',
    timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    type: 'daily',
    start: DateTime.local().plus({ hours: 1 }).startOf('hour').toISO(),
    shiftLength: 1,
    favorite: true,
  })
  const [{ data, error }, commit] = useMutation(mutation)

  if (data?.createRotation) {
    return <Redirect to={`/rotations/${data.createRotation.id}`} />
  }

  return (
    <FormDialog
      title='Create Rotation'
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() =>
        commit({
          input: {
            ...value,
          },
        })
      }
      form={
        <RotationForm
          errors={fieldErrors(error)}
          value={value}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}

export default RotationCreateDialog
