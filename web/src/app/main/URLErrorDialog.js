import React from 'react'
import { Dialog, DialogContent, DialogActions } from '@material-ui/core'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'
import LoadingButton from '../loading/components/LoadingButton'
import { useURLParam } from '../actions'

export default function URLErrorDialog({ onClose }) {
  const errorMessage = useURLParam('errorMessage')
  const errorTitle = useURLParam('errorTitle')

  const open = Boolean(errorMessage) || Boolean(errorTitle)
  const open = Boolean(errorMessage) || Boolean(errorTitle)

  return (
    open && (
      <Dialog
        open={open}
        onClose={() => onClose()}
        aria-labelledby='alert-dialog-title'
        aria-describedby='alert-dialog-description'
      >
        <DialogTitleWrapper id='alert-dialog-title' title={errorTitle} />
        <DialogContent>
          <DialogContentError
            id='alert-dialog-description'
            error={errorMessage}
            noPadding
          />
        </DialogContent>
        <DialogActions>
          <LoadingButton
            buttonText='Okay'
            color='primary'
            onClick={() => onClose()}
          />
        </DialogActions>
      </Dialog>
    )
  )
}
