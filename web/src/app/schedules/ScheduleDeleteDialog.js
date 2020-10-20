import { useQuery, useMutation, gql } from '@apollo/client'
import React from 'react'
import p from 'prop-types'
import { get } from 'lodash-es'
import FormDialog from '../dialogs/FormDialog'
import Spinner from '../loading/components/Spinner'

const query = gql`
  query($id: ID!) {
    schedule(id: $id) {
      id
      name
    }
  }
`
const mutation = gql`
  mutation delete($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function ScheduleDeleteDialog(props) {
  const { data, loading: dataLoading } = useQuery(query, {
    onClose: p.func,
    variables: { id: props.scheduleID },
  })
  const [deleteSchedule, deleteScheduleStatus] = useMutation(mutation, {
    variables: {
      input: [
        {
          type: 'schedule',
          id: props.scheduleID,
        },
      ],
    },
  })

  if (!data && dataLoading) return <Spinner />

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the schedule: ${get(data, 'schedule.name')}`}
      caption='Deleting a schedule will also delete all associated rules and overrides.'
      loading={deleteScheduleStatus.loading}
      errors={deleteScheduleStatus.error ? [deleteScheduleStatus.error] : []}
      onClose={props.onClose}
      onSubmit={() => deleteSchedule()}
    />
  )
}

ScheduleDeleteDialog.propTypes = {
  scheduleID: p.string.isRequired,
  onClose: p.func,
}
