import React, { useEffect, useState } from 'react'
import FormDialog from '../../dialogs/FormDialog'
import UniversalKeyRuleForm from './UniversalKeyRuleForm'
import { gql, useMutation } from 'urql'
import { KeyRuleInput } from '../../../schema'
import { getNotice } from './utils'
import { useErrorConsumer } from '../../util/ErrorConsumer'

interface UniversalKeyRuleCreateDialogProps {
  keyID: string
  onClose: () => void
}

const mutation = gql`
  mutation ($input: UpdateKeyConfigInput!) {
    updateKeyConfig(input: $input)
  }
`

export default function UniversalKeyRuleCreateDialog(
  props: UniversalKeyRuleCreateDialogProps,
): JSX.Element {
  const [value, setValue] = useState<KeyRuleInput>({
    name: '',
    description: '',
    conditionExpr: '',
    continueAfterMatch: false,
    actions: [],
  })
  const [step, setStep] = useState(0)
  const [m, commit] = useMutation(mutation)
  const [hasSubmitted, setHasSubmitted] = useState(0)
  const [hasConfirmed, setHasConfirmed] = useState(false)

  const errs = useErrorConsumer(m.error)
  const errors = errs.remainingLegacyCallback()
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
      title='Create Rule'
      onClose={props.onClose}
      errors={errors}
      onSubmit={() => setHasSubmitted(hasSubmitted + 1)}
      disableSubmit={step < 2 && !hasSubmitted}
      disableNext={step === 2}
      onNext={() => setStep(step + 1)}
      onBack={step > 0 && step <= 2 ? () => setStep(step - 1) : null}
      form={
        <UniversalKeyRuleForm
          value={value}
          onChange={setValue}
          nameError={nameError}
          descriptionError={descError}
          conditionError={conditionError}
          step={step}
          setStep={setStep}
        />
      }
      notices={getNotice(showNotice, hasConfirmed, setHasConfirmed)}
    />
  )
}
