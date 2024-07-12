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

export default function UniversalKeyRuleCreateDialogProps(
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
  const [hasConfirmed, setHasConfirmed] = useState(false)
  const [hasSubmitted, setHasSubmitted] = useState(0)
  const showNotice = hasSubmitted > 0 && value.actions.length === 0

  const errs = useErrorConsumer(m.error)
  const errors = errs.remainingLegacyCallback()
  const [hasErrorAfterSubmit, setHasErrorAfterSubmit] = useState(false)

  const nameError = errs.getErrorByField(/Rules.+\.Name/)
  const descError = errs.getErrorByField(/Rules.+\.Description/)
  const conditionError = errs.getErrorByPath(
    'updateKeyConfig.input.setRule.conditionExpr',
  )

  useEffect(() => {
    if (hasErrorAfterSubmit && (nameError || descError || conditionError)) {
      setStep(0)
      setHasErrorAfterSubmit(false)
    }
  }, [errors, hasErrorAfterSubmit])

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
      if (!res.error) {
        setHasErrorAfterSubmit(false)
        return props.onClose()
      }

      if (nameError || descError || conditionError) {
        setHasErrorAfterSubmit(true)
      }
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
