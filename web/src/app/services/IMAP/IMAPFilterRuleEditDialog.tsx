import React, { useState } from 'react'
import { useQuery, useMutation, gql } from 'urql'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import IMAPFilterRuleForm, { Value } from './IMAPFilterRuleForm'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'

const mutation = gql`
  mutation ($input: UpdateIMAPFilterRuleInput!) {
    updateIMAPFilterRule(input: $input)
  }
`

const query = gql`
  query ($serviceID: ID!) {
    service(id: $serviceID) {
      id
      imapFilterRules {
        id
        name
        enabled
        fromPattern
        subjectPattern
        toPattern
        matchMode
        excludeReplies
      }
    }
  }
`

export default function IMAPFilterRuleEditDialog(props: {
  filterRuleID: string
  serviceID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<Value | null>(null)

  const [{ data, error, fetching }] = useQuery({
    query,
    variables: { serviceID: props.serviceID },
  })
  const [updateStatus, update] = useMutation(mutation)

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const filterRule = data.service.imapFilterRules.find(
    (r: { id: string }) => r.id === props.filterRuleID,
  )

  if (!filterRule) {
    return <GenericError error='Filter rule not found' />
  }

  return (
    <FormDialog
      maxWidth='sm'
      title='Edit IMAP Filter Rule'
      loading={updateStatus.fetching}
      errors={nonFieldErrors(updateStatus.error)}
      onClose={props.onClose}
      onSubmit={() => {
        // Validate that at least one pattern is provided
        const currentValue = value || {
          name: filterRule.name,
          fromPattern: filterRule.fromPattern || '',
          subjectPattern: filterRule.subjectPattern || '',
          toPattern: filterRule.toPattern || '',
          matchMode: filterRule.matchMode.toLowerCase(),
          excludeReplies: filterRule.excludeReplies,
          enabled: filterRule.enabled,
        }

        if (
          !currentValue.fromPattern &&
          !currentValue.subjectPattern &&
          !currentValue.toPattern
        ) {
          return Promise.reject(
            new Error(
              'At least one pattern (From, Subject, or To) must be provided',
            ),
          )
        }

        return update(
          {
            input: {
              id: props.filterRuleID,
              name: currentValue.name,
              // Send empty string to clear, not null (null means "don't update")
              fromPattern: currentValue.fromPattern || '',
              subjectPattern: currentValue.subjectPattern || '',
              toPattern: currentValue.toPattern || '',
              matchMode: currentValue.matchMode,
              excludeReplies: currentValue.excludeReplies,
              enabled: currentValue.enabled,
            },
          },
          { additionalTypenames: ['Service'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }}
      form={
        <IMAPFilterRuleForm
          edit
          errors={fieldErrors(updateStatus.error)}
          disabled={updateStatus.fetching}
          value={
            value || {
              name: filterRule.name,
              fromPattern: filterRule.fromPattern || '',
              subjectPattern: filterRule.subjectPattern || '',
              toPattern: filterRule.toPattern || '',
              matchMode: filterRule.matchMode.toLowerCase(),
              excludeReplies: filterRule.excludeReplies,
              enabled: filterRule.enabled,
            }
          }
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
