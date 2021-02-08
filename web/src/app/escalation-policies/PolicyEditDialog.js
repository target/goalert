import React, { useState } from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'
import p from 'prop-types'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import PolicyForm from './PolicyForm'

const query = gql`
  query($id: ID!) {
    escalationPolicy(id: $id) {
      id
      name
      description
      repeat
    }
  }
`

const mutation = gql`
  mutation($input: UpdateEscalationPolicyInput!) {
    updateEscalationPolicy(input: $input)
  }
`

function PolicyEditDialog(props) {
  const [value, setValue] = useState(null)
  const { data, loading } = useQuery(query, {
    pollInterval: 0,
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

  const [editDialogMutation, editDialogMutationStatus] = useMutation(mutation, {
    variables: {
      input: {
        id: props.escalationPolicyID,
        name: value?.name || defaultValue.name,
        description: value?.description || defaultValue.description,
        repeat: value?.repeat?.value ?? defaultValue.repeat.value,
      },
    },
    onCompleted: props.onClose,
  })
  const fieldErrs = fieldErrors(editDialogMutationStatus.error)

  if (loading && !data.escalationPolicy) return null

  return (
    <FormDialog
      title='Edit Escalation Policy'
      loading={(!data && loading) || editDialogMutationStatus.loading}
      errors={nonFieldErrors(editDialogMutationStatus.error)}
      onClose={props.onClose}
      onSubmit={() => editDialogMutation()}
      form={
        <PolicyForm
          errors={fieldErrs}
          disabled={editDialogMutationStatus.loading}
          value={value || defaultValue}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}

PolicyEditDialog.propTypes = {
  escalationPolicyID: p.string.isRequired,
  onClose: p.func,
}

export default PolicyEditDialog
