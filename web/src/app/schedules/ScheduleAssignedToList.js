import React from 'react'
import { gql } from '@apollo/client'
import Query from '../util/Query'
import FlatList from '../lists/FlatList'
import Card from '@mui/material/Card'
import { useParams } from 'react-router-dom'

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

export default function ScheduleAssignedToList() {
  const { scheduleID } = useParams()
  function renderList({ data }) {
    return (
      <Card style={{ width: '100%' }}>
        <FlatList
          items={data.schedule.assignedTo.map((t) => ({
            title: t.name,
            url: `/escalation-policies/${t.id}`,
          }))}
          emptyMessage='This schedule is not assigned to any escalation policies.'
        />
      </Card>
    )
  }

  return (
    <Query query={query} variables={{ id: scheduleID }} render={renderList} />
  )
}
