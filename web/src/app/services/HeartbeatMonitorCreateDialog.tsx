import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import HeartbeatMonitorForm, { Value } from './HeartbeatMonitorForm'

const createMutation = gql`
  mutation ($input: CreateHeartbeatMonitorInput!) {
    createHeartbeatMonitor(input: $input) {
      id
    }
  }
`

export default function HeartbeatMonitorCreateDialog(props: {
  serviceID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<Value>({ name: '', timeoutMinutes: 15 })
  const [{ error, fetching }, createHeartbeat] = useMutation(createMutation)

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New Heartbeat Monitor'
      loading={fetching}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() =>
        createHeartbeat(
          {
            input: {
              name: value.name,
              timeoutMinutes: value.timeoutMinutes,
              serviceID: props.serviceID,
            },
          },
          { additionalTypenames: ['HeartbeatMonitor'] },
        )
          .then((result) => {
            console.log(result)
          })
          .then(props.onClose)
      }
      form={
        <HeartbeatMonitorForm
          errors={fieldErrors(error).map((f) => ({
            ...f,
            field: f.field === 'timeout' ? 'timeoutMinutes' : f.field,
          }))}
          disabled={fetching}
          value={value}
          onChange={(value: Value) => setValue(value)}
        />
      }
    />
  )
}
