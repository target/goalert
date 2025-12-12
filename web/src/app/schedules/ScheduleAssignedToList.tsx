import React from 'react'
import { useQuery, gql } from 'urql'
import Card from '@mui/material/Card'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import CompList from '../lists/CompList'
import { CompListItemNav } from '../lists/CompListItems'

const query = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      assignedTo {
        id
        type
        name
      }
    }
  }
`

export default function ScheduleAssignedToList(props: {
  scheduleID: string
}): React.JSX.Element {
  const [{ data, fetching, error }] = useQuery({
    query,
    variables: { id: props.scheduleID },
  })

  if (error) {
    return <GenericError error={error.message} />
  }

  if (fetching && !data) {
    return <Spinner />
  }

  return (
    <Card sx={{ width: '100%' }}>
      <CompList emptyMessage='This schedule is not assigned to any escalation policies.'>
        {data.schedule.assignedTo.map((t: { name: string; id: string }) => (
          <CompListItemNav
            key={t.id}
            title={t.name}
            url={`/escalation-policies/${t.id}`}
          />
        ))}
      </CompList>
    </Card>
  )
}
