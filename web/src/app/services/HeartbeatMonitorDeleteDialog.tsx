import React from 'react'
import { useQuery, useMutation, gql } from 'urql'
import { nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const query = gql`
  query ($id: ID!) {
    heartbeatMonitor(id: $id) {
      id
      name
    }
  }
`
const mutation = gql`
  mutation ($id: ID!) {
    deleteAll(input: [{ type: heartbeatMonitor, id: $id }])
  }
`

export default function HeartbeatMonitorDeleteDialog(props: {
  monitorID: string
  onClose: () => void
}): React.JSX.Element {
  const [{ data, fetching, error }] = useQuery({
    query,
    variables: { id: props.monitorID },
  })

  const [deleteHeartbeatStatus, deleteHeartbeat] = useMutation(mutation)

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the heartbeat monitor: ${data.heartbeatMonitor.name}`}
      loading={deleteHeartbeatStatus.fetching}
      errors={nonFieldErrors(deleteHeartbeatStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        deleteHeartbeat(
          { id: props.monitorID },
          { additionalTypenames: ['HeartbeatMonitor'] },
        ).then((res) => {
          if (!res.error) {
            props.onClose()
          }
        })
      }
    />
  )
}
