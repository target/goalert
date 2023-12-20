import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import PolicyStepForm from './PolicyStepForm'
import FormDialog from '../dialogs/FormDialog'
import PolicyStepForm2, { FormValue } from './PolicyStepForm2'

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

function PolicyStepCreateDialog(props: {
  escalationPolicyID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<FormValue | null>(null)
  const defaultValue = {
    actions: [],
    delayMinutes: '15',
  }

  const [createStepStatus, createStep] = useMutation(mutation)

  const { fetching, error } = createStepStatus
  const fieldErrs = fieldErrors(error)

  return (
    <FormDialog
      title='Create Step'
      loading={fetching}
      errors={nonFieldErrors(error)}
      maxWidth='sm'
      onClose={props.onClose}
      onSubmit={() =>
        createStep({
          input: {
            escalationPolicyID: props.escalationPolicyID,
            delayMinutes: parseInt(
              (value && value.delayMinutes) || defaultValue.delayMinutes,
            ),
            targets: (value && value.targets) || defaultValue.targets,
          },
        }).then((result) => {
          if (!result.error) {
            props.onClose()
          }
        })
      }
      form={
        <PolicyStepForm2
          errors={fieldErrs}
          disabled={fetching}
          value={value || defaultValue}
          onChange={(value: Value) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyStepCreateDialog
