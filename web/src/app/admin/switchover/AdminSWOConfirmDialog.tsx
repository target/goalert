import React from 'react'
import FormDialog from '../../dialogs/FormDialog'

export default function AdminSWOConfirmDialog(props: {
  message: string
  onConfirm: () => void
  onClose: () => void
}): JSX.Element {
  return (
    <FormDialog
      title='Continue with switchover?'
      confirm
      subTitle={props.message}
      onClose={props.onClose}
      onSubmit={() => {
        props.onConfirm()
        props.onClose()
      }}
    />
  )
}
