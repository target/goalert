import React from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import { Mutation } from '@apollo/client/react/components'
import { nonFieldErrors } from '../util/errutil'

const mutation = gql`
  mutation ($id: ID!) {
    deleteAll(input: [{ id: $id, type: notificationRule }])
  }
`
export default function UserNotificationRuleDeleteDialog(props) {
  const { ruleID, ...rest } = props
  function renderDialog(commit, { loading, error }) {
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
  return (
    <Mutation mutation={mutation} onCompleted={props.onClose}>
      {(commit, status) => renderDialog(commit, status)}
    </Mutation>
  )
}

UserNotificationRuleDeleteDialog.propTypes = {
  ruleID: p.string.isRequired,
  onClose: p.func,
}
