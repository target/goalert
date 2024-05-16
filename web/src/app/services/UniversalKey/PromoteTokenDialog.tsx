import React, { useState } from 'react'
import {
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  FormControlLabel,
  Grid,
  Typography,
} from '@mui/material'
import { gql, useMutation } from 'urql'
import DialogTitleWrapper from '../../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../../dialogs/components/DialogContentError'

interface PromoteTokenDialogProps {
  keyID: string
  open: boolean
  onClose: () => void
}

const promoteMutation = gql`
  mutation ($id: ID!) {
    promoteSecondaryToken(id: $id)
  }
`

export default function PromoteTokenDialog({
  keyID,
  open,
  onClose,
}: PromoteTokenDialogProps): JSX.Element {
  const [status, commit] = useMutation(promoteMutation)
  const [hasConfirmed, setHasConfirmed] = useState(false)

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitleWrapper
        title='Promote Secondary Token'
        onClose={onClose}
        fullScreen={false}
      />
      <DialogContent>
        <Grid container spacing={2}>
          <Grid item xs={12}>
            <Typography>
              <b>Important note:</b> Promoting this token will delete the
              existing primary token. Any future API requests using the existing
              primary token will fail.
            </Typography>
          </Grid>
          <Grid item xs={12}>
            <FormControlLabel
              control={
                <Checkbox
                  checked={hasConfirmed}
                  onChange={() => setHasConfirmed(!hasConfirmed)}
                />
              }
              label='I acknowledge the impact of this action'
            />
          </Grid>
        </Grid>
      </DialogContent>

      {status.error?.message ? (
        <DialogContentError error={status.error.message} />
      ) : null}

      <DialogActions>
        <Button
          variant='contained'
          disabled={!hasConfirmed}
          onClick={() =>
            commit(
              {
                id: keyID,
              },
              { additionalTypenames: ['IntegrationKey', 'Service'] },
            ).then((res) => {
              if (!res.error) {
                onClose()
              }
            })
          }
        >
          Promote Key
        </Button>
      </DialogActions>
    </Dialog>
  )
}
