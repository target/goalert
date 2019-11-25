import React from 'react'
import p from 'prop-types'
import { Mutation } from 'react-apollo'
import FormDialog from '../dialogs/FormDialog'
import ScheduleRuleForm from './ScheduleRuleForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import gql from 'graphql-tag'
import { startCase } from 'lodash-es'

const mutation = gql`
  mutation($input: ScheduleTargetInput!) {
    updateScheduleTarget(input: $input)
  }
`

export default class ScheduleRuleCreateDialog extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
    targetType: p.oneOf(['rotation', 'user']).isRequired,
    onClose: p.func,
  }

  state = {
    value: {
      targetID: '',
      rules: [
        {
          start: '00:00',
          end: '00:00',
          weekdayFilter: [true, true, true, true, true, true, true],
        },
      ],
    },
  }

  render() {
    return (
      <Mutation mutation={mutation} onCompleted={this.props.onClose}>
        {this.renderDialog}
      </Mutation>
    )
  }

  renderDialog = (commit, status) => {
    return (
      <FormDialog
        onClose={this.props.onClose}
        title={`Add ${startCase(this.props.targetType)} to Schedule`}
        errors={nonFieldErrors(status.error)}
        maxWidth='md'
        onSubmit={() => {
          commit({
            variables: {
              input: {
                target: {
                  type: this.props.targetType,
                  id: this.state.value.targetID,
                },
                scheduleID: this.props.scheduleID,

                rules: this.state.value.rules,
              },
            },
          })
        }}
        form={
          <ScheduleRuleForm
            targetType={this.props.targetType}
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
