import React, { useState } from 'react'
import { Checkbox, FormControlLabel, Typography } from '@mui/material'
import { gql, useMutation } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'

interface PromoteTokenDialogProps {
  keyID: string
  onClose: () => void
}

const mutation = gql`
  mutation ($id: ID!) {
    promoteSecondaryToken(id: $id)
  }
`

export default function PromoteTokenDialog({
  keyID,
  onClose,
}: PromoteTokenDialogProps): React.JSX.Element {
  const [status, commit] = useMutation(mutation)
  const [hasConfirmed, setHasConfirmed] = useState(false)

  return (
    <FormDialog
      title='Promote Secondary Token'
      onClose={onClose}
      errors={nonFieldErrors(status.error)}
      subTitle={
        <Typography>
          <b>Important note:</b> Promoting the secondary token will
          delete/replace the existing primary token. Any future API requests
          using the existing primary token will fail.
        </Typography>
      }
      primaryActionLabel='Promote Key'
      form={
        <FormControlLabel
          control={
            <Checkbox
              checked={hasConfirmed}
              onChange={() => setHasConfirmed(!hasConfirmed)}
            />
          }
          label='I acknowledge the impact of this action'
        />
      }
      disableSubmit={!hasConfirmed}
      onSubmit={() =>
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
    />
  )
}
