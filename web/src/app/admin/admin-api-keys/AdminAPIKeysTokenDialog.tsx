/* eslint-disable prettier/prettier */
import React, { useState, useEffect } from 'react'
import { CreatedGQLAPIKey } from '../../../schema'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import DialogContentText from '@mui/material/DialogContentText'
import DialogTitle from '@mui/material/DialogTitle'
import { Typography } from '@mui/material'
import ContentCopyIcon from '@mui/icons-material/ContentCopy'
import Grid from '@mui/material/Grid'
import IconButton from '@mui/material/IconButton'
import CloseIcon from '@mui/icons-material/Close'

export default function AdminAPIKeysTokenDialog(props: {
  input: CreatedGQLAPIKey
  onTokenDialogClose: (input: boolean) => void
  tokenDialogClose: boolean
}): JSX.Element {
  const [close, onClose] = useState(props.tokenDialogClose)
  const onClickCopy = (): void => {
    navigator.clipboard.writeText(props.input.token)
  }
  const onCloseDialog = (event: any, reason: any): any => {
    if (reason === 'backdropClick' || reason === 'escapeKeyDown') {
      return false;
    }

    onClose(!close)
  }

  const onCloseDialogByButton = (): void => {
    onClose(!close)
  }

  useEffect(() => {
    props.onTokenDialogClose(close)
  })

  return (
    <Dialog
      open={props.tokenDialogClose}
      onClose={onCloseDialog}
      aria-labelledby='api-key-token-dialog'
      aria-describedby='api-key-token-information'
      maxWidth='xl'
      disableEscapeKeyDown
    >
      <DialogTitle id='alert-dialog-api-key-token'>
        <Grid container direction='row' justifyContent='flex-start' display='block'>
          <Grid item style={{ float: 'left' }}>
            <Typography variant='h5'>API Key Token</Typography>
            <Typography variant='subtitle2'>
              <i>
                (Please copy and save the token as this is the only time you'll be able to view it.)
              </i>
            </Typography>
          </Grid>
          <Grid item style={{ float: 'right' }}>
            <IconButton aria-label='close' onClick={onCloseDialogByButton}>
              <CloseIcon />
            </IconButton>
          </Grid>
        </Grid>
      </DialogTitle>
      <DialogContent dividers>
        <DialogContentText id='alert-dialog-api-key-token-content'>
          <Typography sx={{ wordBreak: "break-word" }}>
            {props.input.token}
          </Typography>
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <IconButton aria-label='copy' onClick={onClickCopy}>
          <ContentCopyIcon />
        </IconButton>
      </DialogActions>
    </Dialog>
  )
}
