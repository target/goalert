import React, { useState } from 'react'
import FormDialog from '../../dialogs/FormDialog'
import UniversalKeyRuleForm from './UniversalKeyRuleForm'
import { gql, useMutation } from 'urql'
import { nonFieldErrors } from '../../util/errutil'
import { KeyRuleInput } from '../../../schema'
import { getNotice } from './utils'

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
    id: '',
    name: '',
    description: '',
    conditionExpr: '',
    actions: [],
  })
  const [createStatus, commit] = useMutation(mutation)

  const [hasConfirmed, setHasConfirmed] = useState(false)
  const [hasSubmitted, setHasSubmitted] = useState(false)
  const noActionsNoConf = value.actions.length === 0 && !hasConfirmed

  return (
    <FormDialog
      title='Create Rule'
      maxWidth='lg'
      onClose={props.onClose}
      errors={nonFieldErrors(createStatus.error)}
      onSubmit={() => {
        if (noActionsNoConf) {
          setHasSubmitted(true)
          return
        }

        console.log('submitting')

        return commit(
          {
            input: {
              keyID: props.keyID,
              setRule: {
                name: value.name,
                description: value.description,
                conditionExpr: value.conditionExpr,
                actions: value.actions,
              },
            },
          },
          { additionalTypenames: ['IntegrationKey', 'Service'] },
        ).then((res) => {
          if (!res.error) {
            props.onClose()
          }
        })
      }}
      form={<UniversalKeyRuleForm value={value} onChange={setValue} />}
      notices={getNotice(hasSubmitted, hasConfirmed, setHasConfirmed)}
    />
  )
}
