import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import p from 'prop-types'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import { Redirect } from 'react-router-dom'
import FormDialog from '../dialogs/FormDialog'
import PolicyForm from './PolicyForm'

const mutation = gql`
  mutation($input: CreateEscalationPolicyInput!) {
    createEscalationPolicy(input: $input) {
      id
    }
  }
`

function PolicyCreateDialog(props) {
  const [value, setValue] = useState(null)
  const defaultValue = {
    name: '',
    description: '',
    repeat: { label: '3', value: '3' },
  }
  const [createPolicy, createPolicyStatus] = useMutation(mutation, {
    variables: {
      input: {
        name: (value && value.name) || defaultValue.name,
        description: (value && value.description) || defaultValue.description,
        repeat: (value && value.repeat.value) || defaultValue.repeat.value,
      },
    },
    onCompleted: props.onClose,
  })

  const { loading, data, error } = createPolicyStatus

  if (data && data.createEscalationPolicy) {
    return (
      <Redirect
        push
        to={`/escalation-policies/${data.createEscalationPolicy.id}`}
      />
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
          value={value || defaultValue}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}

PolicyCreateDialog.propTypes = {
  onClose: p.func,
}

export default PolicyCreateDialog
