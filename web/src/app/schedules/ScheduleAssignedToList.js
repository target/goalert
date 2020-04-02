import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import Query from '../util/Query'
import FlatList from '../lists/FlatList'
import Card from '@material-ui/core/Card'

const query = gql`
  query($id: ID!) {
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

export default class ScheduleAssignedToList extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.scheduleID }}
        render={this.renderList}
      />
    )
  }

  renderList({ data }) {
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
}
