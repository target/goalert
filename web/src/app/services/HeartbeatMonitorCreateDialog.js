import React, { useState } from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import HeartbeatMonitorForm from './HeartbeatMonitorForm'

import { useMutation } from '@apollo/react-hooks'

const createMutation = gql`
  mutation($input: CreateHeartbeatMonitorInput!) {
    createHeartbeatMonitor(input: $input) {
      id
    }
  }
`

export default function HeartbeatMonitorCreateDialog(props) {
  const [value, setValue] = useState({ name: '', timeoutMinutes: 15 })
  const [createHeartbeat, { loading, error }] = useMutation(createMutation, {
    variables: {
      input: {
        name: value.name,
        timeoutMinutes: value.timeoutMinutes,
        serviceID: props.serviceID,
      },
    },
  })

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New Heartbeat Monitor'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() => createHeartbeat().then(props.onClose)}
      form={
        <HeartbeatMonitorForm
          errors={fieldErrors(error).map(f => ({
            ...f,
            field: f.field === 'timeout' ? 'timeoutMinutes' : f.field,
          }))}
          disabled={loading}
          value={value}
          onChange={value => setValue(value)}
        />
      }
    />
  )
}

HeartbeatMonitorCreateDialog.propTypes = {
  serviceID: p.string.isRequired,
  onClose: p.func,
}
