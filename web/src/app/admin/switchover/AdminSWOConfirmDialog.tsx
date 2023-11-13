import React from 'react'
import FormDialog from '../../dialogs/FormDialog'

export default function AdminSWOConfirmDialog(props: {
  messages: string[]
  onConfirm: () => void
  onClose: () => void
}): React.ReactNode {
  return (
    <FormDialog
      title='Continue with switchover?'
      confirm
      subTitle='One or more possible problems were detected.'
      onClose={props.onClose}
      onSubmit={() => {
        props.onConfirm()
        props.onClose()
      }}
      form={
        <ul>
          {props.messages.map((m, idx) => {
            return <li key={idx}>{m}</li>
          })}
        </ul>
      }
    />
  )
}
