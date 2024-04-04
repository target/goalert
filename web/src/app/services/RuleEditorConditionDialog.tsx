import React, { useState } from 'react'

import FormDialog from '../dialogs/FormDialog'

export default function RuleEditorConditionDialog(props: {
  expr: string
  onClose: (expr: string | null) => void
}): JSX.Element {
  const [value, setValue] = useState<string>(props.expr)

  return (
    <FormDialog
      maxWidth='sm'
      title='Edit Condition'
      onClose={() => props.onClose(null)}
      onSubmit={() => props.onClose(value)}
      form={
        <textarea
          value={value}
          onChange={(e) => {
            setValue(e.target.value)
          }}
        />
      }
    />
  )
}
