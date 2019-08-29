import React from 'react'
import p from 'prop-types'
import { Mutation } from 'react-apollo'
import FormDialog from '../dialogs/FormDialog'
import ScheduleRuleForm from './ScheduleRuleForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import gql from 'graphql-tag'
import { startCase, pick } from 'lodash-es'
import Query from '../util/Query'

const query = gql`
  query($id: ID!, $tgt: TargetInput!) {
    schedule(id: $id) {
      id
      target(input: $tgt) {
        rules {
          start
          end
          weekdayFilter
        }
      }
    }
  }
`

const mutation = gql`
  mutation($input: ScheduleTargetInput!) {
    updateScheduleTarget(input: $input)
  }
`

export default class ScheduleRuleEditDialog extends React.Component {
  static propTypes = {
    scheduleID: p.string.isRequired,
    target: p.shape({
      type: p.oneOf(['rotation', 'user']).isRequired,
      id: p.string.isRequired,
    }).isRequired,
    onClose: p.func,
  }

  state = {
    value: null,
  }

  shouldComponentUpdate(nextProps, nextState) {
    if (this.state !== nextState) return true

    return false
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.scheduleID, tgt: this.props.target }}
        noPoll
        fetchPolicy={'network-only'}
        render={({ data }) => this.renderMutation(data.schedule.target)}
      />
    )
  }

  renderMutation(data) {
    return (
      <Mutation
        mutation={mutation}
        onCompleted={this.props.onClose}
        refetchQueries={['scheduleRules']}
      >
        {(commit, status) => this.renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  renderDialog(data, commit, status) {
    const defaults = {
      targetID: this.props.target.id,
      rules: data.rules.map(r =>
        pick(r, ['id', 'start', 'end', 'weekdayFilter']),
      ),
    }
    return (
      <FormDialog
        onClose={this.props.onClose}
        title={`Edit Rules for ${startCase(this.props.target.type)}`}
        errors={nonFieldErrors(status.error)}
        maxWidth='md'
        onSubmit={() => {
          if (!this.state.value) {
            // no changes
            this.props.onClose()
            return
          }
          commit({
            variables: {
              input: {
                target: this.props.target,
                scheduleID: this.props.scheduleID,

                rules: this.state.value.rules,
              },
            },
          })
        }}
        form={
          <ScheduleRuleForm
            targetType={this.props.target.type}
            targetDisabled
            scheduleID={this.props.scheduleID}
            disabled={status.loading}
            errors={fieldErrors(status.error)}
            value={this.state.value || defaults}
            onChange={value => this.setState({ value })}
          />
        }
      />
    )
  }
}
