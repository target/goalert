import React from 'react'
import p from 'prop-types'
import { Mutation } from 'react-apollo'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors } from '../util/errutil'
import gql from 'graphql-tag'
import { graphql2Client } from '../apollo'
import { startCase } from 'lodash-es'
import Query from '../util/Query'

const query = gql`
  query($id: ID!, $tgt: TargetInput!) {
    schedule(id: $id) {
      id
      target(input: $tgt) {
        target {
          id
          name
          type
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

export default class ScheduleRuleDeleteDialog extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
    target: p.shape({ id: p.string.isRequired, type: p.string.isRequired })
      .isRequired,
    onClose: p.func,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{
          id: this.props.scheduleID,
          tgt: this.props.target,
        }}
        noPoll
        render={({ data }) => this.renderMutation(data.schedule.target)}
      />
    )
  }
  renderMutation(data) {
    return (
      <Mutation
        mutation={mutation}
        client={graphql2Client}
        onCompleted={this.props.onClose}
        refetchQueries={['scheduleRules']}
        awaitRefetchQueries
      >
        {(commit, status) => this.renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  renderDialog(data, commit, status) {
    return (
      <FormDialog
        onClose={this.props.onClose}
        title={`Remove ${startCase(this.props.target.type)} From Schedule?`}
        subTitle={`This will remove all rules, as well as end any active or future on-call shifts on this schedule for ${
          this.props.target.type
        }: ${data.target.name}.`}
        caption='Overrides will not be affected.'
        confirm
        errors={nonFieldErrors(status.error)}
        onSubmit={() => {
          commit({
            variables: {
              input: {
                target: this.props.target,
                scheduleID: this.props.scheduleID,

                rules: [],
              },
            },
          })
        }}
      />
    )
  }
}
