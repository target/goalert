import React from 'react'
import { useMutation } from '@apollo/client'
import { useQuery, gql } from 'urql'
import { get } from 'lodash'
import FormDialog from '../dialogs/FormDialog'
import Spinner from '../loading/components/Spinner'

const query = gql`
  query ($id: ID!) {
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

export default function ScheduleDeleteDialog(props: {
  onClose: () => void
  scheduleID: string
}): React.ReactNode {
  const [{ data, fetching }] = useQuery({
    query,
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

  if (!data && fetching) return <Spinner />

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
