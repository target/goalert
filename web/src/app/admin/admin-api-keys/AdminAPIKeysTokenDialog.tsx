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
  const onCloseDialog = (): void => {
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
            <IconButton aria-label='close' onClick={onCloseDialog}>
              <CloseIcon />
            </IconButton>
          </Grid>
        </Grid>
      </DialogTitle>
      <DialogContent dividers>
        <DialogContentText id='alert-dialog-api-key-token-content'>
          <Typography sx={{ wordBreak: "break-word" }}>
            {props.input.token}
            eyJhbGciOiJFUzIyNCIsImtleSI6MCwidHlwIjoiSldUIn0.eyJpc3MiOiJnb2FsZXJ0Iiwic3ViIjoiYzY2MDNiNmQtNzc1ZC00ZTc2LThiMzYtOThiMDkwM2NhZjg2IiwiYXVkIjpbImFwaWtleS12MS9ncmFwaHFsLXYxIl0sImV4cCI6MTY5NTM3ODA2MCwibmJmIjoxNjk1MTE4ODA3LCJpYXQiOjE2OTUxMTg4NjcsImp0aSI6ImEyNWM4N2RiLTUyMGUtNGZlMi04MmY5LWFmZmFlNjBmZjhiMCIsInBvbCI6IkE1RGp2TExERjkzaUI4cUpkRHBwTnFUWkw5OUkrMVJRbCtoT2NQNHU2NTA9In0.dXTqhTmKXPM-VVmBelnETs_o-QUxGoltECRZTdOhOLoJZ508WYZNNnJD8qcobNQMDoIsx25v-Yo
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
