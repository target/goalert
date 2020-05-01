import React from 'react'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import ScheduleForm from './ScheduleForm'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import Query from '../util/Query'

const query = gql`
  query($id: ID!) {
    schedule(id: $id) {
      id
      name
      description
      timeZone
    }
  }
`

const mutation = gql`
  mutation($input: UpdateScheduleInput!) {
    updateSchedule(input: $input)
  }
`

export default class ScheduleEditDialog extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: null,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.scheduleID }}
        render={({ data }) => this.renderMutation(data.schedule)}
      />
    )
  }

  renderMutation(data) {
    return (
      <Mutation mutation={mutation} onCompleted={this.props.onClose}>
        {(...args) => this.renderForm(data, ...args)}
      </Mutation>
    )
  }

  renderForm = (data, commit, status) => {
    return (
      <FormDialog
        onClose={this.props.onClose}
        title='Edit Schedule'
        errors={nonFieldErrors(status.error)}
        onSubmit={() =>
          commit({
            variables: {
              input: {
                id: this.props.scheduleID,
                ...this.state.value,
              },
            },
          })
        }
        form={
          <ScheduleForm
            disabled={status.loading}
            errors={fieldErrors(status.error)}
            value={
              this.state.value || {
                name: data.name,
                description: data.description,
                timeZone: data.timeZone,
              }
            }
            onChange={(value) => this.setState({ value })}
          />
        }
      />
    )
  }
}
