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
  const [m, commit] = useMutation(mutation)
  const [hasConfirmed, setHasConfirmed] = useState(false)
  const [hasSubmitted, setHasSubmitted] = useState(false)
  const noActionsNoConf = value.actions.length === 0 && !hasConfirmed

  const errs = useErrorConsumer(m.error)
  const form = (
    <UniversalKeyRuleForm
      value={value}
      onChange={setValue}
      nameError={errs.getErrorByField(/Rules.+\.Name/)}
      descriptionError={errs.getErrorByField(/Rules.+\.Description/)}
      conditionError={errs.getErrorByPath(
        'updateKeyConfig.input.setRule.conditionExpr',
      )}
    />
  )

  return (
    <FormDialog
      title='Create Rule'
      maxWidth='lg'
      onClose={props.onClose}
      errors={errs.remainingLegacy()}
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
      form={form}
      notices={getNotice(hasSubmitted, hasConfirmed, setHasConfirmed)}
    />
  )
}
