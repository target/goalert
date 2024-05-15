import React from 'react'
import FormDialog from '../../dialogs/FormDialog'

interface UniversalKeyRuleCreateDialogProps {
  onClose: () => void
}

export default function UniversalKeyRuleCreateDialogProps(
  props: UniversalKeyRuleCreateDialogProps,
): JSX.Element {
  return (
    <FormDialog
      title='Are you sure?'
      confirm
      onClose={props.onClose}
      onSubmit={props.onClose}
    />
  )
}
