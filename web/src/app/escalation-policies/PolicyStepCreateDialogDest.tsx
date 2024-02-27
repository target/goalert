import React, { useState, useEffect } from 'react'
import { CombinedError, gql, useMutation } from 'urql'
import { splitErrorsByPath } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import PolicyStepFormDest, { FormValue } from './PolicyStepFormDest'
import { errorPaths } from '../users/UserContactMethodFormDest'

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
  const [err, setErr] = useState<CombinedError | null>(null)

  useEffect(() => {
    setErr(null)
  }, [value])

  useEffect(() => {
    setErr(createStepStatus.error || null)
  }, [createStepStatus.error])

  const [formErrors, otherErrs] = splitErrorsByPath(
    err,
    errorPaths('destinationDisplayInfo.input'),
  )

  return (
    <FormDialog
      title='Create Step'
      loading={createStepStatus.fetching}
      errors={otherErrs}
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
          errors={formErrors}
          disabled={createStepStatus.fetching}
          value={value || defaultValue}
          onChange={(value: FormValue) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyStepCreateDialogDest
