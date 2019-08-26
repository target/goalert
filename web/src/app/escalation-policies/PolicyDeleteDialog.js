import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { Redirect } from 'react-router-dom'
import FormDialog from '../dialogs/FormDialog'

const mutation = gql`
  mutation($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default class PolicyDeleteDialog extends React.PureComponent {
  static propTypes = {
    escalationPolicyID: p.string.isRequired,
    onClose: p.func,
  }

  renderDialog = (commit, mutStatus) => {
    const { loading, error, data } = mutStatus
    if (data && data.deleteAll) {
      return <Redirect push to={`/escalation-policies`} />
    }

    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle='You will not be able to delete this policy if it is in use by one or more services.'
        loading={loading}
        errors={error ? [error] : []}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: [
                {
                  type: 'escalationPolicy',
                  id: this.props.escalationPolicyID,
                },
              ],
            },
          })
        }}
      />
    )
  }

  render() {
    return (
      <Mutation
        mutation={mutation}
        awaitRefetchQueries
        refetchQueries={() => ['epsQuery']}
      >
        {(commit, status) => this.renderDialog(commit, status)}
      </Mutation>
    )
  }
}
