import React, { useState, useEffect } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import UniversalKeyRuleForm from './UniversalKeyRuleForm'
import { IntegrationKey, KeyRuleInput } from '../../../schema'
import { getNotice } from './utils'
import { useErrorConsumer } from '../../util/ErrorConsumer'

interface UniversalKeyRuleEditDialogProps {
  keyID: string
  ruleID: string
  onClose: () => void
  default?: boolean
}

const query = gql`
  query UniversalKeyPage($keyID: ID!, $ruleID: ID!) {
    integrationKey(id: $keyID) {
      id
      config {
        oneRule(id: $ruleID) {
          id
          name
          description
          conditionExpr
          continueAfterMatch
          actions {
            dest {
              type
              args
            }
            params
          }
        }
      }
    }
  }
`

const mutation = gql`
  mutation ($input: UpdateKeyConfigInput!) {
    updateKeyConfig(input: $input)
  }
`

export default function UniversalKeyRuleEditDialog(
  props: UniversalKeyRuleEditDialogProps,
): JSX.Element {
  const [q] = useQuery<{
    integrationKey: IntegrationKey
  }>({
    query,
    variables: {
      keyID: props.keyID,
      ruleID: props.ruleID,
    },
  })
  if (q.error) throw q.error

  // shouldn't happen due to suspense
  if (!q.data) throw new Error('failed to load data')

  const rule = q.data.integrationKey.config.oneRule
  if (!rule) throw new Error('rule not found')

  const [value, setValue] = useState<KeyRuleInput>(rule)
  const [step, setStep] = useState(0)
  const [m, commit] = useMutation(mutation)

  const errs = useErrorConsumer(m.error)
  const errors = errs.remainingLegacyCallback()
  const nameError = errs.getErrorByField(/Rules.+\.Name/)
  const descError = errs.getErrorByField(/Rules.+\.Description/)
  const conditionError = errs.getErrorByPath(
    'updateKeyConfig.input.setRule.conditionExpr',
  )

  const [hasConfirmed, setHasConfirmed] = useState(false)
  const [hasSubmitted, setHasSubmitted] = useState(0)
  const showNotice = hasSubmitted > 0 && value.actions.length === 0

  const firstStepErrors = nameError || descError || conditionError
  useEffect(() => {
    if (firstStepErrors) {
      setStep(0)
    }
  }, [errors])

  const handleCommit = () => {
    commit(
      {
        input: {
          keyID: props.keyID,
          setRule: value,
        },
      },
      { additionalTypenames: ['KeyConfig'] },
    ).then((res) => {
      if (!res.error) props.onClose()
    })
  }

  useEffect(() => {
    if (hasSubmitted) {
      if (showNotice && !hasConfirmed) {
        return
      }
      handleCommit()
    }
  }, [hasSubmitted])

  return (
    <FormDialog
      title={props.default ? 'Edit Default Actions' : 'Edit Rule'}
      onClose={props.onClose}
      onSubmit={() => setHasSubmitted(hasSubmitted + 1)}
      disableSubmit={step < 2 && !hasSubmitted}
      disableNext={step === 2 || (step === 0 && firstStepErrors)}
      onNext={() => setStep(step + 1)}
      onBack={step > 0 && step <= 2 ? () => setStep(step - 1) : null}
      errors={errors}
      form={
        <UniversalKeyRuleForm
          value={value}
          onChange={setValue}
          nameError={errs.getErrorByField(/Rules.+\.Name/)}
          descriptionError={errs.getErrorByField(/Rules.+\.Description/)}
          conditionError={errs.getErrorByPath(
            'updateKeyConfig.input.setRule.conditionExpr',
          )}
          step={step}
          setStep={setStep}
        />
      }
      notices={getNotice(showNotice, hasConfirmed, setHasConfirmed)}
    />
  )
}
