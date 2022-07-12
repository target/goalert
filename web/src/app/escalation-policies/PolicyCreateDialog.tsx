import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import PolicyForm from './PolicyForm'
import { Redirect } from 'wouter'

interface Value {
  name: string
  description: string
  repeat: {
    label: string
    value: string
  }
  favorite: boolean
}

const mutation = gql`
  mutation ($input: CreateEscalationPolicyInput!) {
    createEscalationPolicy(input: $input) {
      id
    }
  }
`

function PolicyCreateDialog(props: { onClose: () => void }): JSX.Element {
  const [value, setValue] = useState<Value | null>(null)
  const defaultValue: Value = {
    name: '',
    description: '',
    repeat: { label: '3', value: '3' },
    favorite: true,
  }
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
              name: (value && value.name) || defaultValue.name,
              description:
                (value && value.description) || defaultValue.description,
              repeat:
                (value && value.repeat.value) || defaultValue.repeat.value,
              favorite: true,
            },
          },
          {
            additionalTypenames: ['EscalationPolicy'],
          },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }
      form={
        <PolicyForm
          errors={fieldErrs}
          disabled={fetching}
          value={value || defaultValue}
          onChange={(value: Value) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyCreateDialog
