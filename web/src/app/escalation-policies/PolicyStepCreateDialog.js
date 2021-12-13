import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { gql, useMutation } from '@apollo/client'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import PolicyStepForm from './PolicyStepForm'
import FormDialog from '../dialogs/FormDialog'

const mutation = gql`
  mutation ($input: CreateEscalationPolicyStepInput!) {
    createEscalationPolicyStep(input: $input) {
      id
      delayMinutes
      targets {
        id
        name
        type
      }
    }
  }
`

function PolicyStepCreateDialog(props) {
  const [value, setValue] = useState(null)
  const defaultValue = {
    targets: [],
    delayMinutes: '15',
  }

  const [createStep, createStepStatus] = useMutation(mutation, {
    variables: {
      input: {
        escalationPolicyID: props.escalationPolicyID,
        delayMinutes: parseInt(
          (value && value.delayMinutes) || defaultValue.delayMinutes,
        ),
        targets: (value && value.targets) || defaultValue.targets,
      },
    },
    onCompleted: props.onClose,
  })

  const { loading, error } = createStepStatus
  const fieldErrs = fieldErrors(error)

  return (
    <FormDialog
      title='Create Step'
      loading={loading}
      errors={nonFieldErrors(error)}
      maxWidth='sm'
      onClose={props.onClose}
      onSubmit={() => createStep()}
      form={
        <PolicyStepForm
          errors={fieldErrs}
          disabled={loading}
          value={value || defaultValue}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}

PolicyStepCreateDialog.propTypes = {
  escalationPolicyID: p.string.isRequired,
  onClose: p.func.isRequired,
}

export default PolicyStepCreateDialog
