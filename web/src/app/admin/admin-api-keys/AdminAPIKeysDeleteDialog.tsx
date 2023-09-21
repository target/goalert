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

const deleteGQLAPIKeyQuery = gql`
  mutation DeleteGQLAPIKey($input: string!) {
    deleteGQLAPIKey(input: $input)
  }
`

export default function AdminAPIKeysDeleteDialog(props: {
  apiKey: GQLAPIKey | null
  onClose: (param: boolean) => void
  close: boolean
}): JSX.Element {
  const { apiKey, onClose, close } = props
  const [dialogClose, onDialogClose] = useState(close)
  const handleNo = (): void => {
    console.log('NO...........')
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
  const handleYes = (): void => {
    console.log('YES...........')
    deleteAPIKey({
      variables: {
        input: apiKey?.id,
      },
    }).then((result) => {
      if (!result.errors) {
        return result
      }
    })
  }

  if (error) {
    return <GenericError error={error.message} />
  }

  if (loading && !data) {
    // return <Spinner />
  }

  return (
    <Dialog
      open={close}
      aria-labelledby='delete-api-key-dialog'
      aria-describedby='delete-api-key-dialog'
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
