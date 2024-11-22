import React, { useEffect, useState } from 'react'
import FormDialog from '../../dialogs/FormDialog'
import UniversalKeyRuleForm from './UniversalKeyRuleForm'
import { gql, useMutation, useQuery } from 'urql'
import { IntegrationKey, KeyRuleInput } from '../../../schema'
import { getNotice } from './utils'
import { useErrorConsumer } from '../../util/ErrorConsumer'

interface UniversalKeyRuleDialogProps {
  keyID: string
  onClose: () => void

  ruleID?: string // if present, we are editing
  default?: boolean // used when creating default action
}

const editQuery = gql`
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

export default function UniversalKeyRuleDialog(
  props: UniversalKeyRuleDialogProps,
): JSX.Element {
  const [q] = useQuery<{
    integrationKey: IntegrationKey
  }>({
    query: editQuery,
    variables: {
      keyID: props.keyID,
      ruleID: props.ruleID,
    },
    pause: !props.ruleID,
  })
  if (q.error) throw q.error

  // shouldn't happen due to suspense
  if (props.ruleID && !q.data) throw new Error('failed to load data')

  const rule = q.data?.integrationKey.config.oneRule
  if (props.ruleID && !rule) throw new Error('rule not found')

  const [value, setValue] = useState<KeyRuleInput>({
    id: rule?.id ?? undefined,
    name: rule?.name ?? '',
    description: rule?.description ?? '',
    conditionExpr: rule?.conditionExpr ?? '',
    continueAfterMatch: rule?.continueAfterMatch ?? false,
    actions: rule?.actions ?? [],
  })

  const [step, setStep] = useState(0)
  const [m, commit] = useMutation(mutation)
  const [hasSubmitted, setHasSubmitted] = useState(0)
  const [hasConfirmed, setHasConfirmed] = useState(false)
  const [showNextTooltip, setShowNextTooltip] = useState(false)

  const errs = useErrorConsumer(m.error)
  const unknownErrors = errs.remainingLegacyCallback()
  const nameError = errs.getErrorByField(/Rules.+\.Name/)
  const descError = errs.getErrorByField(/Rules.+\.Description/)
  const conditionError = errs.getErrorByPath(
    'updateKeyConfig.input.setRule.conditionExpr',
  )

  const showNotice = hasSubmitted > 0 && value.actions.length === 0

  useEffect(() => {
    // if no actions notice, user must confirm they want this before submitting
    if ((showNotice && !hasConfirmed) || !hasSubmitted) {
      return
    }

    commit(
      {
        input: {
          keyID: props.keyID,
          setRule: value,
        },
      },
      { additionalTypenames: ['KeyConfig'] },
    ).then((res) => {
      if (res.error) return

      props.onClose()
    })
  }, [hasSubmitted])

  useEffect(() => {
    // showing notice takes precedence
    // don't change steps when this flips to true
    if (showNotice && !hasConfirmed) {
      return
    }

    if (nameError || descError || conditionError) {
      setStep(0)
    }
  }, [nameError, descError, conditionError, showNotice])

  return (
    <FormDialog
      title={
        props.ruleID
          ? props.default
            ? 'Edit Default Actions'
            : 'Edit Rule'
          : 'Create Rule'
      }
      onClose={props.onClose}
      loading={m.fetching}
      maxWidth='md'
      errors={unknownErrors}
      onSubmit={() => setHasSubmitted(hasSubmitted + 1)}
      disableSubmit={step < 2 && !hasSubmitted}
      disableNext={step === 2 || showNextTooltip}
      onNext={() => setStep(step + 1)}
      nextTooltip={
        showNextTooltip
          ? 'Reset or finish adding the current action to continue'
          : null
      }
      onBack={
        step > 0 && step <= 2
          ? () => {
              setShowNextTooltip(false)
              setStep(step - 1)
            }
          : null
      }
      form={
        <UniversalKeyRuleForm
          value={value}
          onChange={setValue}
          nameError={nameError}
          descriptionError={descError}
          conditionError={conditionError}
          step={step}
          setStep={setStep}
          setShowNextTooltip={(bool: boolean) => setShowNextTooltip(bool)}
        />
      }
      notices={getNotice(showNotice, hasConfirmed, setHasConfirmed)}
      PaperProps={{
        sx: {
          minHeight: '500px',
        },
      }}
    />
  )
}
