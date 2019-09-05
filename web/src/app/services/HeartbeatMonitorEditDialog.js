import React, { useState, useEffect } from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'
import FormDialog from '../dialogs/FormDialog'
import HeartbeatMonitorForm from './HeartbeatMonitorForm'
import { useMutation } from '@apollo/react-hooks'

const mutation = gql`
  mutation($input: UpdateHeartbeatMonitorInput!) {
    updateHeartbeatMonitor(input: $input)
  }
`
const query = gql`
  query($id: ID!) {
    heartbeatMonitor(id: $id) {
      id
      name
      timeoutMinutes
    }
  }
`

export default function HeartbeatMonitorEditDialog(props) {
  return (
    <Query
      query={query}
      variables={{ id: props.monitorID }}
      render={({ data }) => (
        <HeartbeatMonitorEditDialogContent
          props={props}
          data={data.heartbeatMonitor}
          refetchQueries={props.refetchQueries}
        />
      )}
    />
  )
}
HeartbeatMonitorEditDialog.propTypes = {
  monitorID: p.string.isRequired,
  refetchQueries: p.arrayOf(p.string),
  onClose: p.func,
}

// TODO: broken out until `useQuery` is built
function HeartbeatMonitorEditDialogContent({ props, data }) {
  const [value, setValue] = useState({
    name: data.name,
    timeoutMinutes: data.timeoutMinutes,
  })
  const [update, { loading, error }] = useMutation(mutation, {
    refetchQueries: props.refetchQueries,
    onCompleted: props.onClose,
    variables: {
      input: { id: props.monitorID, ...value },
    },
  })

  useEffect(() => {
    setValue({ name: data.name, timeoutMinutes: data.timeoutMinutes })
  }, [data.name, data.timeoutMinutes])

  return (
    <FormDialog
      maxWidth='sm'
      title='Edit Heartbeat Monitor'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() => update()}
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
