import React from 'react'
import p from 'prop-types'

import gql from 'graphql-tag'

import {
  Button,
  Dialog,
  DialogActions,
  DialogTitle,
  DialogContent,
  DialogContentText,
} from '@material-ui/core'
import DialogContentError from '../dialogs/components/DialogContentError'
import { useQuery } from '@apollo/react-hooks'

const query = gql`
  query($id: ID!) {
    sendTestStatus(cmID: $id)
  }
`
export default function SendTestDialog(props) {
  const {
    title = 'Test Delivery Status',
    onClose,
    sendTestMutationStatus,
    messageID,
  } = props

  const { data } = useQuery(query, {
    variables: {
      cmID: messageID,
    },
    // skip: sendTestMutationStatus.error || sendTestMutationStatus.loading
  })
  console.log(data)
  return (
    <Dialog open onClose={onClose}>
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>
        <DialogContentText>{data?.sendTestStatus ?? ''}</DialogContentText>
      </DialogContent>
      <DialogContentError
        error={sendTestMutationStatus?.error?.message ?? ''}
      />
      <DialogActions>
        <Button color='primary' variant='contained' onClick={onClose}>
          Done
        </Button>
      </DialogActions>
    </Dialog>
  )
}

SendTestDialog.propTypes = {
  messageID: p.string.isRequired,
  onClose: p.func,
  disclaimer: p.string,
  title: p.string,
  subtitle: p.string,
  sendTestMutationStatus: p.object,
}
