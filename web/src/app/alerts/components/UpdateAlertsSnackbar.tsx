import React from 'react'
import Snackbar from '@mui/material/Snackbar'
import { Alert } from '@mui/material'

interface UpdateAlertsSnackbarProps {
  errorMessage: string
  updateMessage: string
  onClose: () => void
  onExited: () => void
  open: boolean
}

function UpdateAlertsSnackbar(props: UpdateAlertsSnackbarProps): JSX.Element {
  const { errorMessage, updateMessage, onExited, onClose, open } = props
  const err = Boolean(errorMessage)

  return (
    <Snackbar
      autoHideDuration={err ? null : 6000}
      TransitionProps={{
        onExited,
      }}
      onClose={onClose}
      open={open}
    >
      <Alert
        severity={err ? 'error' : 'success'}
        onClose={onClose}
        variant='filled'
      >
        {err ? errorMessage : updateMessage}
      </Alert>
    </Snackbar>
  )
}

export default UpdateAlertsSnackbar
