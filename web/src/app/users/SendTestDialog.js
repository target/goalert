import React from 'react'
import p from 'prop-types'

import gql from 'graphql-tag'
import { useQuery } from '@apollo/react-hooks'
import Spinner from '../loading/components/Spinner'

import {
  Button,
  Dialog,
  DialogActions,
  DialogTitle,
  DialogContent,
  DialogContentText,
} from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import DialogContentError from '../dialogs/components/DialogContentError'

const query = gql`
  query($cmID: ID!) {
    sendTestStatus(cmID: $cmID) {
      details
      status
    }
  }
`
const useStyles = makeStyles({
  statusOk: {
    color: '#218626',
  },
  statusWarn: {
    color: '#867321',
  },
  statusError: {
    color: '#862421',
  },
})

export default function SendTestDialog(props) {
  const {
    title = 'Test Delivery Status',
    onClose,
    sendTestMutationStatus,
    messageID,
    contactMethodType,
    contactMethodFromNumber,
    contactMethodToNumber,
  } = props

  const classes = useStyles()

  const { data, loading, error } = useQuery(query, {
    variables: {
      cmID: messageID,
    },
    skip: sendTestMutationStatus.error || sendTestMutationStatus.loading,
  })

  const details = data?.sendTestStatus?.details ?? ''
  const status = data?.sendTestStatus?.status ?? ''
  const errorMessage =
    (sendTestMutationStatus?.error?.message ?? '') || (error?.message ?? '')

  const getLogStatusClass = (status) => {
    switch (status) {
      case 'OK':
        return classes.statusOk
      case 'ERROR':
        return classes.statusError
      default:
        return classes.statusWarn
    }
  }

  console.log(contactMethodType)

  return (
    <Dialog open onClose={onClose}>
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>
        <DialogContentText>
          GoAlert is sending a {contactMethodType} to {contactMethodToNumber}{' '}
          from {contactMethodFromNumber}
        </DialogContentText>
      </DialogContent>
      {((loading && !details) || sendTestMutationStatus.loading) && (
        <DialogContent>
          <Spinner text='Loading...' />
        </DialogContent>
      )}
      {details && (
        <DialogContent>
          <DialogContentText className={getLogStatusClass(status)}>
            {details}
          </DialogContentText>
        </DialogContent>
      )}
      {errorMessage && <DialogContentError error={errorMessage} />}
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
  contactMethodType: p.string,
  contactMethodToNumber: p.string,
  contactMethodFromNumber: p.string,
}
