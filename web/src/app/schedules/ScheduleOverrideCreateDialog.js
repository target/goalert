import React from 'react'
import p from 'prop-types'
import { Mutation } from 'react-apollo'
import FormDialog from '../dialogs/FormDialog'
import { DateTime } from 'luxon'
import ScheduleOverrideForm from './ScheduleOverrideForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import gql from 'graphql-tag'

const copyText = {
  add: {
    title: 'Temporarily Add a User',
    desc:
      'This will add a new shift for the selected user, while the override is active. Existing shifts will remain unaffected.',
  },
  remove: {
    title: 'Temporarily Remove a User',
    desc:
      'This will remove (or split/shorten) shifts belonging to the selected user, while the override is active.',
  },
  replace: {
    title: 'Temporarily Replace a User',
    desc:
      'This will replace the selected user with another during any existing shifts, while the override is active. No new shifts will be created, only who is on-call will be changed.',
  },
}

const mutation = gql`
  mutation($input: CreateUserOverrideInput!) {
    createUserOverride(input: $input) {
      id
    }
  }
`

export default class ScheduleOverrideCreateDialog extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
    variant: p.oneOf(['add', 'remove', 'replace']).isRequired,
    onClose: p.func,
    defaultValue: p.shape({
      addUserID: p.string,
      removeUserID: p.string,
      start: p.string,
      end: p.string,
    }),
  }

  static defaultProps = {
    defaultValue: {},
  }

  constructor(props) {
    super(props)

    this.state = {
      value: {
        addUserID: '',
        removeUserID: '',
        start: DateTime.local()
          .startOf('hour')
          .toISO(),
        end: DateTime.local()
          .startOf('hour')
          .plus({ hours: 8 })
          .toISO(),
        ...props.defaultValue,
      },
    }
  }

  render() {
    return (
      <Mutation
        mutation={mutation}
        onCompleted={this.props.onClose}
        refetchQueries={[
          'scheduleShifts',
          'scheduleCalendarShifts',
          'scheduleOverrides',
        ]}
      >
        {this.renderDialog}
      </Mutation>
    )
  }

  renderDialog = (commit, status) => {
    return (
      <FormDialog
        onClose={this.props.onClose}
        title={copyText[this.props.variant].title}
        subTitle={copyText[this.props.variant].desc}
        errors={nonFieldErrors(status.error)}
        onSubmit={() =>
          commit({
            variables: {
              input: {
                ...this.state.value,
                scheduleID: this.props.scheduleID,
              },
            },
          })
        }
        form={
          <ScheduleOverrideForm
            add={this.props.variant !== 'remove'}
            remove={this.props.variant !== 'add'}
            scheduleID={this.props.scheduleID}
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
