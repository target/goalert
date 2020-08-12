import React, { useEffect } from 'react'
import p from 'prop-types'

import gql from 'graphql-tag'
import { useQuery, useMutation } from '@apollo/react-hooks'
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
import { textColors } from '../styles/statusStyles'

const query = gql`
  query($id: ID!, $number: String!) {
    userContactMethod(id: $id) {
      type
      formattedValue
      lastTestVerifyAt
      lastTestMessageState {
        details
        status
      }
    }
    phoneNumberInfo(number: $number) {
      formatted
    }
  }
`

const mutation = gql`
  mutation($id: ID!) {
    testContactMethod(id: $id)
  }
`

const useStyles = makeStyles({
  ...textColors,
})

export default function SendTestDialog(props) {
  const classes = useStyles()

  const { title = 'Test Delivery Status', onClose, messageID } = props

  let [contactMethodFromNumber] = useConfigValue('Twilio.FromNumber')

  const [sendTest, sendTestStatus] = useMutation(mutation, {
    variables: {
      id: messageID,
    },
  })

  useEffect(() => {
    sendTest()
  }, [])

  const { data, loading, error } = useQuery(query, {
    variables: {
      id: messageID,
      number: contactMethodFromNumber,
    },
    skip: sendTestStatus.error || sendTestStatus.loading,
  })

  const details = data?.userContactMethod?.lastTestMessageState?.details ?? ''
  const status = data?.userContactMethod?.lastTestMessageState?.status ?? ''
  const contactMethodToNumber = data?.userContactMethod?.formattedValue ?? ''
  const contactMethodType = data?.userContactMethod?.type ?? ''
  const lastTestVerifyAt = data?.userContactMethod?.lastTestVerifyAt ?? ''
  // update from number format
  contactMethodFromNumber = data?.phoneNumberInfo?.formatted ?? ''
  const errorMessage =
    (sendTestStatus?.error?.message ?? '') || (error?.message ?? '')

  const getTestStatusClass = (status) => {
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
      {((loading && !details) || sendTestStatus.loading) && (
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
          <DialogContentText className={getTestStatusClass(status)}>
            {toTitleCase(details)}
          </DialogContentText>
          <DialogContentText>
            Your last test message was sent at {lastTestVerifyAt}
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
}
