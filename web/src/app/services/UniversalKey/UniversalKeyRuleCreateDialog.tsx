import React, { useState } from 'react'
import FormDialog from '../../dialogs/FormDialog'
import UniversalKeyRuleForm from './UniversalKeyRuleForm'

interface UniversalKeyRuleCreateDialogProps {
  onClose: () => void
}

export default function UniversalKeyRuleCreateDialogProps(
  props: UniversalKeyRuleCreateDialogProps,
): JSX.Element {
  const [value, setValue] = useState({ name: '', expr: '' })

  return (
    <FormDialog
      title='Create Rule'
      onClose={props.onClose}
      onSubmit={setValue}
      form={<UniversalKeyRuleForm value={value} onChange={setValue} />}
    />
  )
}
