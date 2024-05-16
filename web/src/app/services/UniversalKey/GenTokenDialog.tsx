import React from 'react'
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  FormHelperText,
  Grid,
  Typography,
} from '@mui/material'

import DialogTitleWrapper from '../../dialogs/components/DialogTitleWrapper'
import CopyText from '../../util/CopyText'
import { gql, useMutation } from 'urql'
import DialogContentError from '../../dialogs/components/DialogContentError'

interface TokenCopyDialogProps {
  keyID: string
  open: boolean
  onClose: () => void
  isSecondary?: boolean
}

const genMutation = gql`
  mutation ($id: ID!) {
    generateKeyToken(id: $id)
  }
`

function GenerateTokenText(): JSX.Element {
  return (
    <React.Fragment>
      <DialogContent>
        <Typography>
          Create an token to externally authenticate with GoAlert when creating
          new alerts
        </Typography>
      </DialogContent>
    </React.Fragment>
  )
}

function CopyTokenText({ token }: { token: string }): JSX.Element {
  return (
    <DialogContent>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <CopyText title={token} value={token} placement='bottom' />
          <FormHelperText>
            Please copy and save the token as this is the only time you'll be
            able to view it.
          </FormHelperText>
        </Grid>
      </Grid>
    </DialogContent>
  )
}

export default function GenTokenDialog({
  keyID,
  open,
  onClose,
  isSecondary,
}: TokenCopyDialogProps): JSX.Element {
  const title = isSecondary ? 'Generate Secondary Token' : 'Generate Token'
  const [status, commit] = useMutation(genMutation)
  const token = status.data?.generateKeyToken

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitleWrapper title={title} onClose={onClose} fullScreen={false} />
      {!token ? <GenerateTokenText /> : <CopyTokenText token={token} />}

      {status.error?.message ? (
        <DialogContentError error={status.error.message} />
      ) : null}

      {!token && (
        <DialogActions>
          <Button
            variant='contained'
            onClick={() =>
              commit(
                {
                  id: keyID,
                },
                { additionalTypenames: ['IntegrationKey', 'Service'] },
              )
            }
          >
            Generate
          </Button>
        </DialogActions>
      )}
    </Dialog>
  )
}
