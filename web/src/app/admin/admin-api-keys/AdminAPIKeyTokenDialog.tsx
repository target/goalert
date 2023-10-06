/* eslint-disable prettier/prettier */
import React from 'react'
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

export default function AdminAPIKeyTokenDialog(props: {
  value: CreatedGQLAPIKey
  onClose: () => void
}): JSX.Element {
  const {onClose, value} = props
  const onClickCopy = (): void => {
    navigator.clipboard.writeText(props.value.token)
  }
  // handles onclose dialog for the token dialog, rejects close for backdropclick or escapekeydown actions
  const onCloseDialog = (
    event: object,
    reason: string,
  ): boolean | undefined => {
    if (reason === 'backdropClick' || reason === 'escapeKeyDown') {
      return false
    }

    onClose()
  }
  // handles close dialog button action
  const onCloseDialogByButton = (): void => {
    onClose()
  }

  return (
    <Dialog
      open
      onClose={onCloseDialog}
      aria-labelledby='api-key-token-dialog'
      aria-describedby='api-key-token-information'
      maxWidth='xl'
      disableEscapeKeyDown
    >
      <DialogTitle id='alert-dialog-api-key-token'>
        <Grid
          container
          direction='row'
          justifyContent='flex-start'
          display='block'
        >
          <Grid item style={{ float: 'left' }}>
            <Typography variant='h5'>API Key Token</Typography>
            <Typography variant='subtitle2'>
              <i>
                (Please copy and save the token as this is the only time you'll
                be able to view it.)
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
          <Typography sx={{ wordBreak: 'break-word' }}>
            {value.token}
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
