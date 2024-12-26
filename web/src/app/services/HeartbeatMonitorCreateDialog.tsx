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
}): React.JSX.Element {
  const [value, setValue] = useState<Value>({
    name: '',
    timeoutMinutes: 15,
    additionalDetails: '',
  })
  const [createHeartbeatStatus, createHeartbeat] = useMutation(createMutation)

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New Heartbeat Monitor'
      loading={createHeartbeatStatus.fetching}
      errors={nonFieldErrors(createHeartbeatStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        createHeartbeat(
          {
            input: {
              name: value.name,
              timeoutMinutes: value.timeoutMinutes,
              serviceID: props.serviceID,
              additionalDetails: value.additionalDetails,
            },
          },
          { additionalTypenames: ['HeartbeatMonitor', 'Service'] },
        ).then((result) => {
          if (!result.error) {
            props.onClose()
          }
        })
      }
      form={
        <HeartbeatMonitorForm
          errors={fieldErrors(createHeartbeatStatus.error).map((f) => ({
            ...f,
            field: f.field === 'timeout' ? 'timeoutMinutes' : f.field,
          }))}
          disabled={createHeartbeatStatus.fetching}
          value={value}
          onChange={(value: Value) => setValue(value)}
        />
      }
    />
  )
}
