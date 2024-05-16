import React from 'react'
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
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

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitleWrapper
        title='Promote Secondary Token'
        onClose={onClose}
        fullScreen={false}
      />
      <DialogContent>
        <Typography>Generate a token for use</Typography>
      </DialogContent>

      {status.error?.message ? (
        <DialogContentError error={status.error.message} />
      ) : null}

      <DialogActions>
        <Button
          variant='contained'
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
