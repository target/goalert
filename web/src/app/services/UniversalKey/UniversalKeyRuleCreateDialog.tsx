import React, { useState } from 'react'
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
  const [hasSubmitted, setHasSubmitted] = useState(false)
  const noActionsNoConf = value.actions.length === 0 && !hasConfirmed

  const errs = useErrorConsumer(m.error)

  return (
    <FormDialog
      title='Create Rule'
      onClose={props.onClose}
      errors={errs.remainingLegacyCallback()}
      onSubmit={() => {
        if (noActionsNoConf) {
          setHasSubmitted(true)
          return
        }

        return commit(
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
      }}
      onNext={step >= 0 && step < 2 ? () => setStep(step + 1) : null}
      onBack={step > 0 && step <= 2 ? () => setStep(step - 1) : null}
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
      notices={getNotice(hasSubmitted, hasConfirmed, setHasConfirmed)}
    />
  )
}
