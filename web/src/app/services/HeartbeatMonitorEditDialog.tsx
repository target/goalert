import React, { useState } from 'react'
import { useQuery, useMutation, gql } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import HeartbeatMonitorForm, { Value } from './HeartbeatMonitorForm'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const mutation = gql`
  mutation ($input: UpdateHeartbeatMonitorInput!) {
    updateHeartbeatMonitor(input: $input)
  }
`
const query = gql`
  query ($id: ID!) {
    heartbeatMonitor(id: $id) {
      id
      name
      timeoutMinutes
    }
  }
`

export default function HeartbeatMonitorEditDialog(props: {
  monitorID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<Value | null>(null)

  const [{ data: qData, error: qError, fetching: qFetching }] = useQuery({
    query,
    variables: { id: props.monitorID },
  })
  const [{ error, fetching }, update] = useMutation(mutation)

  if (qFetching && !qData) return <Spinner />
  if (qError) return <GenericError error={qError.message} />

  return (
    <FormDialog
      maxWidth='sm'
      title='Edit Heartbeat Monitor'
      loading={fetching}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() =>
        update(
          { input: { id: props.monitorID, ...value } },
          { additionalTypenames: ['HeartbeatMonitor'] },
        ).then(props.onClose)
      }
      form={
        <HeartbeatMonitorForm
          errors={fieldErrors(error).map((f) => ({
            ...f,
            field: f.field === 'timeout' ? 'timeoutMinutes' : f.field,
          }))}
          disabled={fetching}
          value={
            value || {
              name: qData.heartbeatMonitor.name,
              timeoutMinutes: qData.heartbeatMonitor.timeoutMinutes,
            }
          }
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
