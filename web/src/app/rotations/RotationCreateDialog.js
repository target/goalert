import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import p from 'prop-types'
import { Redirect } from 'react-router'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import RotationForm from './RotationForm'
import { DateTime } from 'luxon'

const mutation = gql`
  mutation($input: CreateRotationInput!) {
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
const initialValue = {
  name: '',
  description: '',
  timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  type: 'daily',
  start: DateTime.local().plus({ hours: 1 }).startOf('hour').toISO(),
  shiftLength: 1,
  favorite: true,
}

const RotationCreateDialog = (props) => {
  const [value, setValue] = useState(initialValue)
  const [createRotationMutation, { loading, data, error }] = useMutation(
    mutation,
    {
      variables: {
        input: {
          ...value,
        },
      },
    },
  )

  if (data?.createRotation) {
    return <Redirect push to={`/rotations/${data.createRotation.id}`} />
  }

  return (
    <FormDialog
      title='Create Rotation'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() => createRotationMutation()}
      form={
        <RotationForm
          errors={fieldErrors(error)}
          disabled={loading}
          value={value}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}

RotationCreateDialog.propTypes = {
  onClose: p.func.isRequired,
}

export default RotationCreateDialog
