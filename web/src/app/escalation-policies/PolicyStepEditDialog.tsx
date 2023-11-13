import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import PolicyStepForm from './PolicyStepForm'
import FormDialog from '../dialogs/FormDialog'
import { UpdateEscalationPolicyStepInput } from '../../schema'

interface PolicyStepEditDialogProps {
  escalationPolicyID: string
  onClose: () => void
  step: UpdateEscalationPolicyStepInput
}

const mutation = gql`
  mutation ($input: UpdateEscalationPolicyStepInput!) {
    updateEscalationPolicyStep(input: $input)
  }
`

function PolicyStepEditDialog(props: PolicyStepEditDialogProps): React.ReactNode {
  const [value, setValue] = useState<UpdateEscalationPolicyStepInput | null>(
    null,
  )

  const defaultValue = {
    targets: props.step?.targets?.map(({ id, type }) => ({ id, type })),
    delayMinutes: props.step?.delayMinutes?.toString(),
  }

  const [editStepMutationStatus, editStepMutation] = useMutation(mutation)

  const { fetching, error } = editStepMutationStatus
  const fieldErrs = fieldErrors(error)

  return (
    <FormDialog
      title='Edit Step'
      loading={fetching}
      errors={nonFieldErrors(error)}
      maxWidth='sm'
      onClose={props.onClose}
      onSubmit={() =>
        editStepMutation(
          {
            input: {
              id: props.step.id,
              delayMinutes:
                (value && value.delayMinutes) || defaultValue.delayMinutes,
              targets: (value && value.targets) || defaultValue.targets,
            },
          },
          { additionalTypenames: ['EscalationPolicy'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }
      form={
        <PolicyStepForm
          errors={fieldErrs}
          disabled={fetching}
          value={value || defaultValue}
          onChange={(value: UpdateEscalationPolicyStepInput) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyStepEditDialog
