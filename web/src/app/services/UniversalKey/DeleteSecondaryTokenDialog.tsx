import React, { useState } from 'react'
import { Checkbox, FormControlLabel, Typography } from '@mui/material'
import { gql, useMutation } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'

interface DeleteSecondaryTokenDialogProps {
  keyID: string
  onClose: () => void
}

const mutation = gql`
  mutation ($id: ID!) {
    deleteSecondaryToken(id: $id)
  }
`

export default function DeleteSecondaryTokenDialog({
  keyID,
  onClose,
}: DeleteSecondaryTokenDialogProps): JSX.Element {
  const [status, commit] = useMutation(mutation)
  const [hasConfirmed, setHasConfirmed] = useState(false)

  return (
    <FormDialog
      title='Delete Secondary Token'
      onClose={onClose}
      errors={nonFieldErrors(status.error)}
      subTitle={
        <Typography>
          <b>Important note:</b> Deleting the secondary authentication token
          will cause any future API requests using this token to fail.
        </Typography>
      }
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
