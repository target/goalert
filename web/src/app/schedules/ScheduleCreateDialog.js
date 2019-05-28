import React from 'react'
import FormDialog from '../dialogs/FormDialog'
import ScheduleForm from './ScheduleForm'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { graphql2Client } from '../apollo'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import { Redirect } from 'react-router'

const mutation = gql`
  mutation($input: CreateScheduleInput!) {
    createSchedule(input: $input) {
      id
      name
      description
      timeZone
    }
  }
`

export default class ScheduleCreateDialog extends React.PureComponent {
  state = {
    value: {
      name: '',
      description: '',
      timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    },
  }
  render() {
    return (
      <Mutation mutation={mutation} client={graphql2Client}>
        {this.renderForm}
      </Mutation>
    )
  }
  renderForm = (commit, status) => {
    if (status.data && status.data.createSchedule) {
      return (
        <Redirect push to={`/schedules/${status.data.createSchedule.id}`} />
      )
    }
    return (
      <FormDialog
        onClose={this.props.onClose}
        title='Create New Schedule'
        errors={nonFieldErrors(status.error)}
        onSubmit={() =>
          commit({
            variables: {
              input: {
                ...this.state.value,
                targets: [
                  {
                    target: { type: 'user', id: '__current_user' },
                    rules: [{}],
                  },
                ],
              },
            },
          })
        }
        form={
          <ScheduleForm
            disabled={status.loading}
            errors={fieldErrors(status.error)}
            value={this.state.value}
            onChange={value => this.setState({ value })}
          />
        }
      />
    )
  }
}
