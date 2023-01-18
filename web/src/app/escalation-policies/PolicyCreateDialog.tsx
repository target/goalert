import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
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

function PolicyCreateDialog(props: { onClose: () => void }): JSX.Element {
  const defaultValue = {
    name: '',
    description: '',
    repeat: { label: '3', value: '3' },
    favorite: true,
  }
  const [value, setValue] = useState<PolicyFormValue>(defaultValue)
  const [createPolicy, createPolicyStatus] = useMutation(mutation, {
    variables: {
      input: {
        name: value && value.name,
        description: value && value.description,
        repeat: value && value.repeat.value,
        favorite: true,
      },
    },
  })

  const { loading, data, error } = createPolicyStatus

  if (data && data.createEscalationPolicy) {
    return (
      <Redirect to={`/escalation-policies/${data.createEscalationPolicy.id}`} />
    )
  }

  const fieldErrs = fieldErrors(error)

  return (
    <FormDialog
      title='Create Escalation Policy'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() => createPolicy()}
      form={
        <PolicyForm
          errors={fieldErrs}
          disabled={loading}
          value={value}
          onChange={(value: PolicyFormValue) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyCreateDialog
