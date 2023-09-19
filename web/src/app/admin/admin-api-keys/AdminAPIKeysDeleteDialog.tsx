import React, { useState } from 'react'
import { GQLAPIKey } from '../../../schema'
import Button from '@mui/material/Button'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import DialogContentText from '@mui/material/DialogContentText'
import DialogTitle from '@mui/material/DialogTitle'

/**
const mutation = gql`
  mutation delete($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
` */

export default function AdminAPIKeysDeleteDialog(props: {
  apiKey: GQLAPIKey | null
  onClose: (param: boolean) => void
  close: boolean
}): JSX.Element {
  const { apiKey, onClose, close } = props
  const [dialogClose, onDialogClose] = useState(close)
  const handleNo = (): void => {
    onClose(false)
    onDialogClose(!dialogClose)
  }

  const handleYes = (): void => {
    onClose(false)
    onDialogClose(!dialogClose)
  }

  return (
    <Dialog
      open={close}
      aria-labelledby='delete-api-key-dialog'
      aria-describedby='delete-api-key-dialog'
    >
      <DialogTitle id='delete-api-key-title'>DELETE API KEY</DialogTitle>
      <DialogContent>
        <DialogContentText id='delete-api-key-content'>
          Are you sure you want to delete the API KEY {apiKey?.name}?
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleNo}>NO</Button>
        <Button onClick={handleYes}>YES</Button>
      </DialogActions>
    </Dialog>
  )
}
