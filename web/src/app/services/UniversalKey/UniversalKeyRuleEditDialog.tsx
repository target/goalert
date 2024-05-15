import React, { useState } from 'react'
import FormDialog from '../../dialogs/FormDialog'
import UniversalKeyRuleForm from './UniversalKeyRuleForm'

interface UniversalKeyRuleEditDialogProps {
  onClose: () => void
}

export default function UniversalKeyRuleCreateDialogProps(
  props: UniversalKeyRuleEditDialogProps,
): JSX.Element {
  const [value, setValue] = useState({ name: '', expr: '' })

  return (
    <FormDialog
      title='Edit Rule'
      onClose={props.onClose}
      onSubmit={setValue}
      form={<UniversalKeyRuleForm value={value} onChange={setValue} />}
    />
  )
}
