import React from 'react'
import p from 'prop-types'
import { Mutation } from 'react-apollo'
import { DateTime } from 'luxon'
import FormDialog from '../dialogs/FormDialog'
import ScheduleOverrideForm from './ScheduleOverrideForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import gql from 'graphql-tag'
import Query from '../util/Query'
import { connect } from 'react-redux'
import { urlParamSelector } from '../selectors'

const query = gql`
  query($id: ID!) {
    userOverride(id: $id) {
      id
      start
      end
      target {
        id
      }
      addUser {
        id
      }
      removeUser {
        id
      }
    }
  }
`
const mutation = gql`
  mutation($input: UpdateUserOverrideInput!) {
    updateUserOverride(input: $input)
  }
`

@connect(state => ({ zone: urlParamSelector(state)('tz', 'local') }))
export default class ScheduleOverrideEditDialog extends React.PureComponent {
  static propTypes = {
    overrideID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: null,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.overrideID }}
        noPoll
        fetchPolicy='network-only'
        render={({ data }) => this.renderMutation(data.userOverride)}
      />
    )
  }

  renderMutation(data) {
    return (
      <Mutation
        mutation={mutation}
        onCompleted={this.props.onClose}
        refetchQueries={['scheduleShifts', 'scheduleOverrides']}
      >
        {(commit, status) => this.renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  getDefaultValue(data) {
    const value = {
      start: DateTime.fromISO(data.start, { zone: this.props.zone }),
      end: DateTime.fromISO(data.end, { zone: this.props.zone }),
    }

    value.addUserID = data.addUser ? data.addUser.id : ''
    value.removeUserID = data.removeUser ? data.removeUser.id : ''

    return value
  }

  renderDialog(data, commit, status) {
    const defaultValue = this.getDefaultValue(data)

    return (
      <FormDialog
        onClose={this.props.onClose}
        title='Edit Schedule Override'
        errors={nonFieldErrors(status.error)}
        onSubmit={() => {
          if (this.state.value === null) {
            this.props.onClose()
            return
          }

          commit({
            variables: {
              input: {
                ...this.state.value,
                id: this.props.overrideID,
                start: this.state.value.start.toISO(),
                end: this.state.value.end.toISO(),
              },
            },
          })
        }}
        form={
          <ScheduleOverrideForm
            add={Boolean(data.addUser)}
            remove={Boolean(data.removeUser)}
            scheduleID={data.target.id}
            disabled={status.loading}
            errors={fieldErrors(status.error)}
            value={this.state.value || defaultValue}
            onChange={value => this.setState({ value })}
          />
        }
      />
    )
  }
}
