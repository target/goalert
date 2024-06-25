import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import { splitErrorsByPath } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import PolicyStepForm, { FormValue } from './PolicyStepForm'
import {
  Destination,
  EscalationPolicy,
  UpdateEscalationPolicyStepInput,
} from '../../schema'
import { getNotice } from './utils'

interface PolicyStepEditDialogProps {
  escalationPolicyID: string
  onClose: () => void
  stepID: string
  disablePortal?: boolean
}

const mutation = gql`
  mutation UpdateEPStep($input: UpdateEscalationPolicyStepInput!) {
    updateEscalationPolicyStep(input: $input)
  }
`

const query = gql`
  query GetEPStep($id: ID!) {
    escalationPolicy(id: $id) {
      id
      steps {
        id
        delayMinutes
        actions {
          type
          args
        }
      }
    }
  }
`

export default function PolicyStepEditDialog(
  props: PolicyStepEditDialogProps,
): React.ReactNode {
  const [stepQ] = useQuery<{ escalationPolicy: EscalationPolicy }>({
    query,
    variables: { id: props.escalationPolicyID },
  })
  const step = stepQ.data?.escalationPolicy.steps.find(
    (s) => s.id === props.stepID,
  )

  if (!step) throw new Error('Step not found')

  const [value, setValue] = useState<FormValue>({
    actions: (step.actions || []).map((a: Destination) => ({
      // remove extraneous fields
      type: a.type,
      args: a.args,
    })),
    delayMinutes: step.delayMinutes,
  })

  const [editStepStatus, editStep] = useMutation(mutation)

  const [hasSubmitted, setHasSubmitted] = useState(false)
  const [hasConfirmed, setHasConfirmed] = useState(false)
  const noActionsNoConf = value.actions.length === 0 && !hasConfirmed

  // Edit dialog has no errors to be handled by the form:
  // - actions field has it's own validation
  // - errors on existing actions are not handled specially, and just display in the dialog (i.e., duplicates)
  // - the delay field has no validation, and is automatically clamped to the min/max values by the backend
  const [, errs] = splitErrorsByPath(editStepStatus.error, [])

  return (
    <FormDialog
      title='Edit Step'
      loading={editStepStatus.fetching}
      errors={errs}
      disablePortal={props.disablePortal}
      maxWidth='sm'
      onClose={props.onClose}
      onSubmit={() => {
        if (noActionsNoConf) {
          setHasSubmitted(true)
          return
        }

        return editStep(
          {
            input: {
              id: props.stepID,
              delayMinutes: +value.delayMinutes,
              actions: value.actions,
            } satisfies UpdateEscalationPolicyStepInput,
          },
          { additionalTypenames: ['EscalationPolicy'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }}
      form={
        <PolicyStepForm
          disabled={editStepStatus.fetching}
          value={value}
          onChange={(value: FormValue) => setValue(value)}
        />
      }
      notices={getNotice(hasSubmitted, hasConfirmed, setHasConfirmed)}
    />
  )
}
