import React, { useState } from 'react'
import { GQLAPIKey } from '../../../schema'
import Button from '@mui/material/Button'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import DialogContentText from '@mui/material/DialogContentText'
import DialogTitle from '@mui/material/DialogTitle'
import { gql, useMutation } from '@apollo/client'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'

// query for deleting API Key which accepts API Key ID
const deleteGQLAPIKeyQuery = gql`
  mutation DeleteGQLAPIKey($id: ID!) {
    deleteGQLAPIKey(id: $id)
  }
`

export default function AdminAPIKeysDeleteDialog(props: {
  apiKey: GQLAPIKey | null
  onClose: (param: boolean) => void
  close: boolean
}): JSX.Element {
  const { apiKey, onClose, close } = props
  const [dialogClose, onDialogClose] = useState(close)
  // handles the no confirmation option for delete API Key transactions
  const handleNo = (): void => {
    onClose(false)
    onDialogClose(!dialogClose)
  }
  const [deleteAPIKey, deleteAPIKeyStatus] = useMutation(deleteGQLAPIKeyQuery, {
    onCompleted: (data) => {
      if (data.deleteGQLAPIKey) {
        onClose(false)
        onDialogClose(!dialogClose)
      }
    },
  })
  const { loading, data, error } = deleteAPIKeyStatus
  // handles the yes confirmation option for delete API Key transactions
  const handleYes = (): void => {
    deleteAPIKey({
      variables: {
        id: apiKey?.id,
      },
    })
  }

  if (error) {
    return <GenericError error={error.message} />
  }

  if (loading && !data) {
    return <Spinner />
  }

  return (
    <Dialog
      open={close}
      onClose={onClose}
      aria-labelledby='delete-api-key-dialog-label'
      aria-describedby='delete-api-key-dialog-desc'
    >
      <DialogTitle id='delete-api-key-title'>DELETE API KEY</DialogTitle>
      <DialogContent>
        <DialogContentText id='delete-api-key-content'>
          Are you sure you want to delete the API KEY {apiKey?.name}?
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleNo}>NO</Button>
        <Button onClick={handleYes}>YES</Button>
      </DialogActions>
    </Dialog>
  )
}
