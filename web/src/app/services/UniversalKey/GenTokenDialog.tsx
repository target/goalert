import React from 'react'
import { FormHelperText, Grid, Typography } from '@mui/material'
import CopyText from '../../util/CopyText'
import { gql, useMutation } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'

interface TokenCopyDialogProps {
  keyID: string
  onClose: () => void
}

const mutation = gql`
  mutation ($id: ID!) {
    generateKeyToken(id: $id)
  }
`

function GenerateTokenText(): React.JSX.Element {
  return (
    <Typography>
      Generates an auth bearer token that can immediately be used for this
      integration.
    </Typography>
  )
}

function CopyTokenText({ token }: { token: string }): React.JSX.Element {
  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <CopyText title={token} value={token} placement='bottom' />
        <FormHelperText>
          Please copy and save the token as this is the only time you'll be able
          to view it.
        </FormHelperText>
      </Grid>
    </Grid>
  )
}

export default function GenTokenDialog({
  keyID,
  onClose,
}: TokenCopyDialogProps): React.JSX.Element {
  const [status, commit] = useMutation(mutation)
  const token = status.data?.generateKeyToken

  return (
    <FormDialog
      alert={!!token}
      title={token ? 'Auth Token' : 'Generate Auth Token'}
      onClose={onClose}
      errors={nonFieldErrors(status.error)}
      disableBackdropClose
      primaryActionLabel={!token ? 'Generate' : 'Done'}
      form={!token ? <GenerateTokenText /> : <CopyTokenText token={token} />}
      onSubmit={() =>
        !token
          ? commit(
              {
                id: keyID,
              },
              { additionalTypenames: ['IntegrationKey', 'Service'] },
            )
          : onClose()
      }
    />
  )
}
