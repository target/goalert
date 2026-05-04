import React from 'react'
import { useMutation, gql } from 'urql'
import { nonFieldErrors } from '../../util/errutil'

import FormDialog from '../../dialogs/FormDialog'

const mutation = gql`
  mutation ($id: ID!) {
    deleteIMAPFilterRule(id: $id)
  }
`

export default function IMAPFilterRuleDeleteDialog(props: {
  filterRuleID: string
  filterRuleName: string
  onClose: () => void
}): JSX.Element {
  const [deleteStatus, deleteRule] = useMutation(mutation)

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the IMAP filter rule: ${props.filterRuleName}`}
      loading={deleteStatus.fetching}
      errors={nonFieldErrors(deleteStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        deleteRule(
          { id: props.filterRuleID },
          { additionalTypenames: ['Service'] },
        ).then((res) => {
          if (!res.error) {
            props.onClose()
          }
        })
      }
    />
  )
}
