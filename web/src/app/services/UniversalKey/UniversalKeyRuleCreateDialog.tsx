import React, { useState } from 'react'
import FormDialog from '../../dialogs/FormDialog'
import UniversalKeyRuleForm from './UniversalKeyRuleForm'
import { gql, useMutation } from 'urql'
import { nonFieldErrors } from '../../util/errutil'
import { KeyRuleInput } from '../../../schema'
import { Checkbox, FormControlLabel } from '@mui/material'

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
      onClose={props.onClose}
      onSubmit={() => {
        if (noActionsNoConf) {
          setHasSubmitted(true)
          return
        }

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
        ).then(() => {
          props.onClose()
        })
      }}
      form={<UniversalKeyRuleForm value={value} onChange={setValue} />}
      errors={nonFieldErrors(createStatus.error)}
      notices={
        hasSubmitted
          ? [
              {
                type: 'WARNING',
                message: 'No actions',
                details:
                  'If you submit with no actions created, nothing will happen on this step',
                action: (
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={hasConfirmed}
                        onChange={() => setHasConfirmed(!hasConfirmed)}
                      />
                    }
                    label='I acknowledge the impact of this'
                  />
                ),
              },
            ]
          : []
      }
    />
  )
}
