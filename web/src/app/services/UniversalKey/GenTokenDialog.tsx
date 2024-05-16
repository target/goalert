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

function GenerateToken({ generate }: { generate: () => void }): JSX.Element {
  return (
    <React.Fragment>
      <DialogContent>
        <Typography>Generate a token for external use with GoAlert</Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={generate}>Generate</Button>
      </DialogActions>
    </React.Fragment>
  )
}

function CopyToken({ token }: { token: string }): JSX.Element {
  return (
    <DialogContent>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Typography sx={{ mb: 3 }}>
            <CopyText title={token} value={token} placement='bottom' />
          </Typography>
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
      {!token ? (
        <GenerateToken
          generate={() => {
            commit(
              {
                id: keyID,
              },
              { additionalTypenames: ['IntegrationKey', 'Service'] },
            )
          }}
        />
      ) : (
        <CopyToken token={token} />
      )}
      {status.error?.message ? (
        <DialogContentError error={status.error.message} />
      ) : null}
    </Dialog>
  )
}
