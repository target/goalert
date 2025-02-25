import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import PolicyForm, { PolicyFormValue } from './PolicyForm'
import Spinner from '../loading/components/Spinner'

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
}): React.JSX.Element {
  const [value, setValue] = useState<PolicyFormValue | null>(null)
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

  if (fetching && !data?.escalationPolicy) return <Spinner />

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
              description: value?.description,
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
          onChange={(value: PolicyFormValue) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyEditDialog
