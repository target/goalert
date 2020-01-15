import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'

import FormDialog from '../dialogs/FormDialog'
import { useMutation } from '@apollo/react-hooks'

const query = gql`
  query($id: ID!) {
    heartbeatMonitor(id: $id) {
      id
      name
    }
  }
`
const mutation = gql`
  mutation($id: ID!) {
    deleteAll(input: [{ type: heartbeatMonitor, id: $id }])
  }
`

export default function HeartbeatMonitorDeleteDialog(props) {
  const [deleteHeartbeat, { loading, error }] = useMutation(mutation, {
    variables: { id: props.monitorID },
  })

  function renderQuery() {
    return (
      <Query
        query={query}
        variables={{ id: props.monitorID }}
        render={({ data }) => renderDialog(data.heartbeatMonitor.name)}
      />
    )
  }

  function renderDialog(name) {
    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete the heartbeat monitor: ${name}`}
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={props.onClose}
        onSubmit={() => deleteHeartbeat().then(props.onClose)}
      />
    )
  }

  return renderQuery()
}

HeartbeatMonitorDeleteDialog.propTypes = {
  monitorID: p.string.isRequired,
  onClose: p.func,
}
