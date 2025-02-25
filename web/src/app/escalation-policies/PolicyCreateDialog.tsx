import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import PolicyForm, { PolicyFormValue } from './PolicyForm'
import { Redirect } from 'wouter'

const mutation = gql`
  mutation ($input: CreateEscalationPolicyInput!) {
    createEscalationPolicy(input: $input) {
      id
    }
  }
`

function PolicyCreateDialog(props: { onClose: () => void }): React.JSX.Element {
  const defaultValue = {
    name: '',
    description: '',
    repeat: { label: '3', value: '3' },
    favorite: true,
  }
  const [value, setValue] = useState<PolicyFormValue>(defaultValue)
  const [createPolicyStatus, createPolicy] = useMutation(mutation)

  const { fetching, data, error } = createPolicyStatus

  if (data && data.createEscalationPolicy) {
    return (
      <Redirect to={`/escalation-policies/${data.createEscalationPolicy.id}`} />
    )
  }

  const fieldErrs = fieldErrors(error)

  return (
    <FormDialog
      title='Create Escalation Policy'
      loading={fetching}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() =>
        createPolicy(
          {
            input: {
              name: value && value.name,
              description: value && value.description,
              repeat: value && value.repeat.value,
              favorite: true,
            },
          },
          { additionalTypenames: ['EscalationPolicyConnection'] },
        )
      }
      form={
        <PolicyForm
          errors={fieldErrs}
          disabled={fetching}
          value={value}
          onChange={(value: PolicyFormValue) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyCreateDialog
