import React, { useEffect, useState } from 'react'
import { CombinedError, gql, useMutation, useQuery } from 'urql'
import { splitErrorsByPath } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import PolicyStepFormDest, { FormValue } from './PolicyStepFormDest'
import { errorPaths } from '../users/UserContactMethodFormDest'
import {
  Destination,
  EscalationPolicy,
  FieldValuePair,
  UpdateEscalationPolicyStepInput,
} from '../../schema'

interface PolicyStepEditDialogDestProps {
  escalationPolicyID: string
  onClose: () => void
  stepID: string
}

const mutation = gql`
  mutation ($input: UpdateEscalationPolicyStepInput!) {
    updateEscalationPolicyStep(input: $input)
  }
`

const query = gql`
  query GetEscalationPolicyStep($id: ID!) {
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

function PolicyStepEditDialogDest(
  props: PolicyStepEditDialogDestProps,
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
      values: a.values.map((v: FieldValuePair) => ({
        fieldID: v.fieldID,
        value: v.value,
      })),
    })),
    delayMinutes: step.delayMinutes,
  })

  const [editStepStatus, editStep] = useMutation(mutation)
  const [stepErr, setStepErr] = useState<CombinedError | null>(null)

  const [formErrors, otherErrs] = splitErrorsByPath(
    stepErr,
    errorPaths('destinationDisplayInfo.input'),
  )

  useEffect(() => {
    setStepErr(null)
  }, [value])

  useEffect(() => {
    setStepErr(editStepStatus.error || null)
  }, [editStepStatus.error])

  return (
    <FormDialog
      title='Edit Step'
      loading={editStepStatus.fetching}
      errors={otherErrs}
      maxWidth='sm'
      onClose={props.onClose}
      onSubmit={() =>
        editStep(
          {
            input: {
              id: props.stepID,
              delayMinutes: value.delayMinutes,
              actions: value.actions,
            } satisfies UpdateEscalationPolicyStepInput,
          },
          { additionalTypenames: ['EscalationPolicy'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }
      form={
        <PolicyStepFormDest
          errors={formErrors}
          disabled={editStepStatus.fetching}
          value={value}
          onChange={(value: FormValue) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyStepEditDialogDest
