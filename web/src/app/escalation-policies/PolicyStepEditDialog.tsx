import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import { EscalationPolicy } from '../../schema'
import PolicyStepForm2, { FormValue } from './PolicyStepForm2'

interface PolicyStepEditDialogProps {
  escalationPolicyID: string
  stepID: string
  onClose: () => void
}

const query = gql`
  query PolicyStepEditDialog($id: ID!) {
    escalationPolicy(id: $id) {
      id
      steps {
        id
        delayMinutes
        actions {
          type
          values {
            fieldID
            value
          }
        }
      }
    }
  }
`

const mutation = gql`
  mutation ($input: UpdateEscalationPolicyStepInput!) {
    updateEscalationPolicyStep(input: $input)
  }
`

function PolicyStepEditDialog(props: PolicyStepEditDialogProps): JSX.Element {
  const [{ data, error }] = useQuery<{ escalationPolicy: EscalationPolicy }>({
    query,
    variables: { id: props.escalationPolicyID },
  })
  if (error) throw error
  const step = data?.escalationPolicy.steps?.find(
    (step) => step.id === props.stepID,
  )
  if (!step) throw new Error('Step not found')

  const [value, setValue] = useState<FormValue>({
    delayMinutes: step.delayMinutes,
    actions: step.actions,
  })

  const [status, editStepMutation] = useMutation(mutation)

  return (
    <FormDialog
      title='Edit Step'
      loading={status.fetching}
      errors={nonFieldErrors(error)}
      maxWidth='sm'
      onClose={props.onClose}
      onSubmit={() =>
        editStepMutation(
          {
            input: {
              id: step.id,
              ...value,
            },
          },
          { additionalTypenames: ['EscalationPolicy'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }
      form={
        <PolicyStepForm2
          errors={fieldErrors(status.error)}
          disabled={status.fetching}
          value={value}
          onChange={(value: FormValue) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyStepEditDialog
