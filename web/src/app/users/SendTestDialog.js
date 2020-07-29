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
import toTitleCase from '../util/toTitleCase.ts'
import { useConfigValue } from '../util/RequireConfig'

const query = gql`
  query($cmID: ID!, $number: String!) {
    sendTestStatus(cmID: $cmID) {
      details
      status
    }
    userContactMethod(id: $cmID) {
      type
      formattedValue
    }
    phoneNumberInfo(number: $number) {
      formatted
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
  const classes = useStyles()

  const {
    title = 'Test Delivery Status',
    onClose,
    sendTestMutationStatus,
    messageID,
  } = props

  let [contactMethodFromNumber] = useConfigValue('Twilio.FromNumber')

  const { data, loading, error } = useQuery(query, {
    variables: {
      cmID: messageID,
      number: contactMethodFromNumber,
    },
    skip: sendTestMutationStatus.error || sendTestMutationStatus.loading,
  })

  const details = data?.sendTestStatus?.details ?? ''
  const status = data?.sendTestStatus?.status ?? ''
  const contactMethodToNumber = data?.userContactMethod?.formattedValue ?? ''
  const contactMethodType = data?.userContactMethod?.type ?? ''
  // update from number format
  contactMethodFromNumber = data?.phoneNumberInfo?.formatted ?? ''
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

  return (
    <Dialog open onClose={onClose}>
      <DialogTitle>{title}</DialogTitle>
      {((loading && !details) || sendTestMutationStatus.loading) && (
        <DialogContent>
          <Spinner text='Loading...' />
        </DialogContent>
      )}
      {details && (
        <DialogContent>
          <DialogContentText>
            GoAlert is sending a {contactMethodType} to {contactMethodToNumber}{' '}
            from {contactMethodFromNumber}
          </DialogContentText>
          <DialogContentText className={getLogStatusClass(status)}>
            {toTitleCase(details)}
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
}
