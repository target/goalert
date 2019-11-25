import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import FormDialog from '../dialogs/FormDialog'
import { Mutation } from 'react-apollo'
import { nonFieldErrors } from '../util/errutil'

const mutation = gql`
  mutation($id: ID!) {
    deleteAll(input: [{ id: $id, type: notificationRule }])
  }
`
export default class UserNotificationRuleDeleteDialog extends React.PureComponent {
  static propTypes = {
    ruleID: p.string.isRequired,
  }

  render() {
    return (
      <Mutation mutation={mutation} onCompleted={this.props.onClose}>
        {(commit, status) => this.renderDialog(commit, status)}
      </Mutation>
    )
  }

  renderDialog(commit, { loading, error }) {
    const { ruleID, ...rest } = this.props
    return (
      <FormDialog
        title='Are you sure?'
        confirm
        loading={loading}
        errors={nonFieldErrors(error)}
        subTitle='This will delete the notification rule.'
        onSubmit={() => commit({ variables: { id: ruleID } })}
        {...rest}
      />
    )
  }
}
