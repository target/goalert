import React, { useState, ReactNode } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import PolicyForm from './PolicyForm'

interface Value {
  name: string
  description: string
  repeat: {
    label: string
    value: string
  }
}

const query = gql`
  query ($id: ID!) {
    escalationPolicy(id: $id) {
      id
      name
      description
      repeat
    }
  }
`

const mutation = gql`
  mutation ($input: UpdateEscalationPolicyInput!) {
    updateEscalationPolicy(input: $input)
  }
`

function PolicyEditDialog(props: {
  escalationPolicyID: string
  onClose: () => void
}): ReactNode {
  const [value, setValue] = useState<Value | null>(null)
  const [{ data, fetching }] = useQuery({
    query,
    variables: { id: props.escalationPolicyID },
  })

  const defaultValue = {
    id: props.escalationPolicyID,
    name: data?.escalationPolicy?.name,
    description: data?.escalationPolicy?.description,
    repeat: {
      label: data?.escalationPolicy?.repeat.toString(),
      value: data?.escalationPolicy?.repeat.toString(),
    },
  }

  const [editDialogMutationStatus, editDialogMutation] = useMutation(mutation)
  const fieldErrs = fieldErrors(editDialogMutationStatus.error)

  if (fetching && !data?.escalationPolicy) return null

  return (
    <FormDialog
      title='Edit Escalation Policy'
      loading={(!data && fetching) || editDialogMutationStatus.fetching}
      errors={nonFieldErrors(editDialogMutationStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        editDialogMutation(
          {
            input: {
              id: props.escalationPolicyID,
              name: value?.name || defaultValue.name,
              description: value?.description || defaultValue.description,
              repeat: value?.repeat?.value ?? defaultValue.repeat.value,
            },
          },
          { additionalTypenames: ['EscalationPolicy'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }
      form={
        <PolicyForm
          errors={fieldErrs}
          disabled={editDialogMutationStatus.fetching}
          value={value || defaultValue}
          onChange={(value: Value) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyEditDialog
