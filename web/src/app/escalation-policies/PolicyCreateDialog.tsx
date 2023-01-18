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
  const [value, setValue] = useState<PolicyFormValue>({
    name: '',
    description: '',
    repeat: { label: '3', value: '3' },
    favorite: true,
  })
  const [createPolicy, { loading, data, error }] = useMutation(mutation, {
    variables: {
      input: {
        name: value.name,
        description: value.description,
        repeat: value.repeat?.value ?? 3,
        favorite: true,
      },
    },
  })

  if (data.createEscalationPolicy) {
    return (
      <Redirect to={`/escalation-policies/${data.createEscalationPolicy.id}`} />
    )
  }

  return (
    <FormDialog
      title='Create Escalation Policy'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() => createPolicy()}
      form={
        <PolicyForm
          errors={fieldErrors(error)}
          disabled={loading}
          value={value}
          onChange={(value: PolicyFormValue) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyCreateDialog
