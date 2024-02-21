import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import PolicyStepFormDest, { FormValue } from './PolicyStepFormDest'

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

function PolicyStepCreateDialogDest(props: {
  escalationPolicyID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<FormValue | null>(null)
  const defaultValue = {
    actions: [],
    delayMinutes: 15,
  }

  const [createStepStatus, createStep] = useMutation(mutation)

  const { fetching, error } = createStepStatus
  const fieldErrs = fieldErrors(error)

  return (
    <FormDialog
      title='Create Step'
      loading={fetching}
      errors={nonFieldErrors(error) || fieldErrs}
      maxWidth='sm'
      onClose={props.onClose}
      onSubmit={() => {
        createStep({
          input: {
            escalationPolicyID: props.escalationPolicyID,
            delayMinutes:
              (value && value.delayMinutes) || defaultValue.delayMinutes,
            actions: (value && value.actions) || defaultValue.actions,
          },
        }).then((result) => {
          if (!result.error) {
            props.onClose()
          }
        })
      }}
      form={
        <PolicyStepFormDest
          disabled={fetching}
          value={value || defaultValue}
          onChange={(value: FormValue) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyStepCreateDialogDest
