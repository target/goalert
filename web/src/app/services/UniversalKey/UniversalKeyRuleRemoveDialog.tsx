import React from 'react'
import FormDialog from '../../dialogs/FormDialog'
import { gql, useMutation } from 'urql'
import { nonFieldErrors } from '../../util/errutil'

interface UniversalKeyRuleCreateDialogProps {
  keyID: string
  ruleID: string
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
  const [removeStatus, commit] = useMutation(mutation)

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      onClose={props.onClose}
      onSubmit={() =>
        commit(
          {
            input: {
              keyID: props.keyID,
              deleteRule: props.ruleID,
            },
          },
          { additionalTypenames: ['KeyConfig'] },
        ).then(() => {
          props.onClose()
        })
      }
      errors={nonFieldErrors(removeStatus.error)}
    />
  )
}
