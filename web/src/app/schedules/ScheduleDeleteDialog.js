import React from 'react'
import p from 'prop-types'

import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { nonFieldErrors } from '../util/errutil'
import { Redirect } from 'react-router'
import Query from '../util/Query'

import FormDialog from '../dialogs/FormDialog'

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

export default class ScheduleDeleteDialog extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    deleteEP: true,
  }

  render() {
    return (
      <Query
        noPoll
        client={graphql2Client}
        query={query}
        variables={{ id: this.props.scheduleID }}
        render={({ data }) => this.renderMutation(data.schedule)}
      />
    )
  }

  renderMutation(data) {
    return (
      <Mutation client={graphql2Client} mutation={mutation}>
        {(commit, status) => this.renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  renderDialog(data, commit, mutStatus) {
    const { loading, error, data: mutData } = mutStatus
    if (mutData && mutData.deleteAll) {
      return <Redirect push to={`/schedules`} />
    }

    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete the schedule: ${data.name}`}
        caption='Deleting a schedule will also delete all associated rules and overrides.'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: [{ type: 'schedule', id: this.props.scheduleID }],
            },
          })
        }}
      />
    )
  }
}
