import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import gql from 'graphql-tag'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import PolicyStepForm from './PolicyStepForm'
import FormDialog from '../dialogs/FormDialog'
import { useURLParam } from '../actions'
import { useMutation } from '@apollo/react-hooks'

const mutation = gql`
  mutation($input: UpdateEscalationPolicyStepInput!) {
    updateEscalationPolicyStep(input: $input)
  }
`

function PolicyStepEditDialog(props) {
  const [value, setValue] = useState(null)

  const [errorMessage] = useURLParam('errorMessage', null)
  const [errorTitle] = useURLParam('errorTitle', null)

  const defaultValue = {
    targets: props.step.targets.map(({ id, type }) => ({ id, type })),
    delayMinutes: props.step.delayMinutes.toString(),
  }

  const [editStepMutation, editStepMutationStatus] = useMutation(mutation, {
    variables: {
      input: {
        id: props.step.id,
        delayMinutes:
          (value && value.delayMinutes) || defaultValue.delayMinutes,
        targets: (value && value.targets) || defaultValue.targets,
      },
    },
    onCompleted: props.onClose,
  })

  const { loading, error } = editStepMutationStatus
  const fieldErrs = fieldErrors(error)

  // don't render dialog if slack redirect returns with an error
  if (Boolean(errorMessage) || Boolean(errorTitle)) {
    return null
  }

  return (
    <FormDialog
      title='Edit Step'
      loading={loading}
      errors={nonFieldErrors(error)}
      maxWidth='sm'
      onClose={props.onClose}
      onSubmit={() => editStepMutation()}
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

PolicyStepEditDialog.propTypes = {
  escalationPolicyID: p.string.isRequired,
  onClose: p.func.isRequired,
  step: p.shape({
    id: p.string.isRequired,
    // number from backend, string from textField
    delayMinutes: p.oneOfType([p.number, p.string]).isRequired,
    targets: p.arrayOf(
      p.shape({
        id: p.string.isRequired,
        name: p.string.isRequired,
        type: p.string.isRequired,
      }),
    ).isRequired,
  }),
}

export default PolicyStepEditDialog
