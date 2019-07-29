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
  mutation($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function HeartbeatMonitorDeleteDialog(props) {
  const [deleteHeartbeat, { loading, error }] = useMutation(mutation, {
    refetchQueries: ['monitorQuery'],
    awaitRefetchQueries: true,
    variables: { id: props.heartbeatID },
    data: {
      id: props.heartbeatID,
    },
  })

  function renderQuery() {
    return (
      <Query
        query={query}
        variables={{ id: props.heartbeatID }}
        render={({ data }) => renderDialog(data.heartbeatMonitor.name)}
      />
    )
  }

  function renderDialog(name) {
    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete the heartbeat: ${name}`}
        caption=''
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={props.onClose}
        onSubmit={() => {
          const input = [
            {
              type: 'heartbeatMonitor',
              id: props.heartbeatID,
            },
          ]
          return deleteHeartbeat({
            variables: {
              input,
            },
          }).then(props.onClose)
        }}
      />
    )
  }

  return renderQuery()
}

HeartbeatMonitorDeleteDialog.propTypes = {
  heartbeatID: p.string.isRequired,
  onClose: p.func,
}
