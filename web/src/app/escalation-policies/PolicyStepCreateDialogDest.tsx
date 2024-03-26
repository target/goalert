import React, { useState, useEffect } from 'react'
import { CombinedError, gql, useMutation } from 'urql'
import { splitErrorsByPath } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import PolicyStepFormDest, { FormValue } from './PolicyStepFormDest'
import { errorPaths } from '../users/UserContactMethodFormDest'

const mutation = gql`
  mutation createEscalationPolicyStep(
    $input: CreateEscalationPolicyStepInput!
  ) {
    createEscalationPolicyStep(input: $input) {
      id
    }
  }
`

function PolicyStepCreateDialogDest(props: {
  escalationPolicyID: string
  disablePortal?: boolean
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<FormValue>({
    actions: [],
    delayMinutes: 15,
  })

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
      disablePortal={props.disablePortal}
      title='Create Step'
      loading={createStepStatus.fetching}
      errors={otherErrs}
      maxWidth='sm'
      onClose={props.onClose}
      onSubmit={() => {
        createStep(
          {
            input: {
              escalationPolicyID: props.escalationPolicyID,
              delayMinutes: +value.delayMinutes,
              actions: value.actions,
            },
          },
          { additionalTypenames: ['EscalationPolicy'] },
        ).then((result) => {
          if (result.error) return

          props.onClose()
        })
      }}
      form={
        <PolicyStepFormDest
          errors={formErrors}
          disabled={createStepStatus.fetching}
          value={value}
          onChange={(value: FormValue) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyStepCreateDialogDest
