import React from 'react'
import p from 'prop-types'
import { Mutation } from 'react-apollo'
import FormDialog from '../dialogs/FormDialog'
import ScheduleRuleForm from './ScheduleRuleForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import gql from 'graphql-tag'
import { startCase } from 'lodash-es'
import { connect } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { DateTime } from 'luxon'

const mutation = gql`
  mutation($input: ScheduleTargetInput!) {
    updateScheduleTarget(input: $input)
  }
`
@connect(state => ({ zone: urlParamSelector(state)('tz', 'local') }))
export default class ScheduleRuleCreateDialog extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
    targetType: p.oneOf(['rotation', 'user']).isRequired,
    onClose: p.func,
  }

  constructor(props) {
    super(props)

    // only care about hour and minute, but it needs to be a parsable date for the time picker
    const zeroZero = DateTime.fromFormat('00:00', 'HH:mm', { zone: props.zone })

    this.state = {
      value: {
        targetID: '',
        rules: [
          {
            start: zeroZero,
            end: zeroZero,
            weekdayFilter: [true, true, true, true, true, true, true],
          },
        ],
      },
    }
  }

  render() {
    return (
      <Mutation
        mutation={mutation}
        onCompleted={this.props.onClose}
        refetchQueries={['scheduleRules']}
      >
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

                rules: this.state.value.rules.map(r => ({
                  ...r,
                  start: r.start.toFormat('HH:mm'),
                  end: r.end.toFormat('HH:mm'),
                })),
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
