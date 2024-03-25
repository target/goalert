import React, { useEffect, useState } from 'react'
import { CombinedError, gql, useClient, useMutation } from 'urql'
import { splitErrorsByPath } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import { UpdateEscalationPolicyStepInput } from '../../schema'
import PolicyStepFormDest, { FormValue } from './PolicyStepFormDest'
import { errorPaths } from '../users/UserContactMethodFormDest'

interface PolicyStepEditDialogDestProps {
  escalationPolicyID: string
  onClose: () => void
  step: UpdateEscalationPolicyStepInput
}

const mutation = gql`
  mutation ($input: UpdateEscalationPolicyStepInput!) {
    updateEscalationPolicyStep(input: $input)
  }
`

const query = gql`
  query GetEscalationPolicyStep($input: ID!) {
    escalationPolicy(id: $input) {
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

interface EscalationPolicyStep {
  actions: Destination[]
  id: string
  delayMinutes: number
}

interface Destination {
  type: string
  values: FieldValuePair[]
}

interface FieldValuePair {
  fieldID: string
  value: string
}

function PolicyStepEditDialogDest(
  props: PolicyStepEditDialogDestProps,
): JSX.Element {
  const defaultValue: FormValue = {
    actions: props.step.actions ?? [],
    delayMinutes: props.step.delayMinutes ?? 0,
  }

  const [value, setValue] = useState<FormValue | null>(null)

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

  const validationClient = useClient()

  useEffect(() => {
    validationClient
      .query(query, {
        input: props.escalationPolicyID,
      })
      .toPromise()
      .then((res) => {
        const matchedStep = res.data.escalationPolicy.steps.find(
          (step: EscalationPolicyStep) => step.id === props.step.id,
        )
        setValue(matchedStep)
      })
  }, [editStepStatus.stale])

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
              id: props.step.id,
              delayMinutes:
                (value && value.delayMinutes) || defaultValue.delayMinutes,
              actions: (value && value.actions) || defaultValue.actions,
            },
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
          value={value || defaultValue}
          onChange={(value: FormValue) => setValue(value)}
        />
      }
    />
  )
}

export default PolicyStepEditDialogDest
