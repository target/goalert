import React, { useState } from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import { Redirect } from 'react-router'
import { Mutation } from '@apollo/client/react/components'
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

  const renderDialog = (commit, status) => {
    const { loading } = status
    if (status.data && status.data.createRotation) {
      return (
        <Redirect push to={`/rotations/${status.data.createRotation.id}`} />
      )
    }

    return (
      <FormDialog
        title='Create Rotation'
        loading={loading}
        errors={nonFieldErrors(status.error)}
        onClose={props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
                timeZone: value.timeZone,
                ...value,
              },
            },
          })
        }}
        form={
          <RotationForm
            errors={fieldErrors(status.error)}
            disabled={status.loading}
            value={value}
            onChange={(value) => setValue(value)}
          />
        }
      />
    )
  }

  return (
    <Mutation mutation={mutation}>
      {(commit, status) => renderDialog(commit, status)}
    </Mutation>
  )
}

RotationCreateDialog.propTypes = {
  onClose: p.func.isRequired,
}

export default RotationCreateDialog
