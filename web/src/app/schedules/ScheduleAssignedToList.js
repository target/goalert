import React from 'react'
import { gql } from '@apollo/client'
import Query from '../util/Query'
import FlatList from '../lists/FlatList'
import Card from '@material-ui/core/Card'
import p from 'prop-types'

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

export default function ScheduleAssignedToList(props) {
  ScheduleAssignedToList.propTypes = {
    scheduleID: p.string.isRequired,
  }

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
    <Query
      query={query}
      variables={{ id: props.scheduleID }}
      render={renderList}
    />
  )
}
