import { gql } from '@apollo/client'
import React from 'react'
import p from 'prop-types'
import { Mutation } from 'react-apollo'
import { nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'
import FormDialog from '../dialogs/FormDialog'

const query = gql`
  query($id: ID!) {
    escalationPolicy(id: $id) {
      id
      steps {
        id
      }
    }
  }
`

const mutation = gql`
  mutation($input: UpdateEscalationPolicyInput!) {
    updateEscalationPolicy(input: $input)
  }
`

export default class PolicyStepDeleteDialog extends React.PureComponent {
  static propTypes = {
    escalationPolicyID: p.string.isRequired,
    stepID: p.string.isRequired,
    onClose: p.func,
  }

  renderMutation(data) {
    return (
      <Mutation
        mutation={mutation}
        onCompleted={this.props.onClose}
        update={(cache) => {
          const { escalationPolicy } = cache.readQuery({
            query,
            variables: { id: this.props.escalationPolicyID },
          })
          cache.writeQuery({
            query,
            variables: { id: data.serviceID },
            data: {
              escalationPolicy: {
                ...escalationPolicy,
                steps: (escalationPolicy.steps || []).filter(
                  (step) => step.id !== this.props.stepID,
                ),
              },
            },
          })
        }}
      >
        {(commit, status) => this.renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  renderDialog(data, commit, mutStatus) {
    const { loading, error } = mutStatus

    // get array of step ids without the step to delete
    const sids = data.steps.map((s) => s.id)
    const toDel = sids.indexOf(this.props.stepID)
    sids.splice(toDel, 1)

    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={
          'This will delete step #' +
          (data.steps.map((s) => s.id).indexOf(this.props.stepID) + 1) +
          ' on this escalation policy.'
        }
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
                id: data.id,
                stepIDs: sids,
              },
            },
          })
        }}
      />
    )
  }

  render() {
    return (
      <Query
        noPoll
        query={query}
        variables={{ id: this.props.escalationPolicyID }}
        render={({ data }) => this.renderMutation(data.escalationPolicy)}
      />
    )
  }
}
