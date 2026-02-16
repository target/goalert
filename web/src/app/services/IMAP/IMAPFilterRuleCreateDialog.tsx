import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'

import FormDialog from '../../dialogs/FormDialog'
import IMAPFilterRuleForm, { Value } from './IMAPFilterRuleForm'

const createMutation = gql`
  mutation ($input: CreateIMAPFilterRuleInput!) {
    createIMAPFilterRule(input: $input) {
      id
      name
    }
  }
`

export default function IMAPFilterRuleCreateDialog(props: {
  serviceID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<Value>({
    name: '',
    fromPattern: '',
    subjectPattern: '',
    toPattern: '',
    matchMode: 'contains',
    excludeReplies: true,
  })
  const [createStatus, createRule] = useMutation(createMutation)

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New IMAP Filter Rule'
      loading={createStatus.fetching}
      errors={nonFieldErrors(createStatus.error)}
      onClose={props.onClose}
      onSubmit={() => {
        // Validate that at least one pattern is provided
        if (!value.fromPattern && !value.subjectPattern && !value.toPattern) {
          return Promise.reject(
            new Error(
              'At least one pattern (From, Subject, or To) must be provided',
            ),
          )
        }

        return createRule(
          {
            input: {
              serviceID: props.serviceID,
              name: value.name,
              fromPattern: value.fromPattern || null,
              subjectPattern: value.subjectPattern || null,
              toPattern: value.toPattern || null,
              matchMode: value.matchMode,
              excludeReplies: value.excludeReplies,
            },
          },
          { additionalTypenames: ['Service'] },
        ).then((result) => {
          if (!result.error) {
            props.onClose()
          }
        })
      }}
      form={
        <IMAPFilterRuleForm
          errors={fieldErrors(createStatus.error)}
          disabled={createStatus.fetching}
          value={value}
          onChange={(value: Value) => setValue(value)}
        />
      }
    />
  )
}
