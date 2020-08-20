import React, { useEffect, useState, MouseEvent } from 'react'

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
import toTitleCase from '../util/toTitleCase'
import DialogContentError from '../dialogs/components/DialogContentError'
import { useConfigValue } from '../util/RequireConfig'
import { textColors } from '../styles/statusStyles'

import { DateTime } from 'luxon'

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

export default function SendTestDialog(
  props: SendTestDialogProps,
): JSX.Element {
  const classes = useStyles()

  const { title = 'Test Delivery Status', onClose, messageID } = props

  let [contactMethodFromNumber] = useConfigValue('Twilio.FromNumber')

  const [now] = useState(DateTime.local())

  const [sendTest, sendTestStatus] = useMutation(mutation, {
    variables: {
      id: messageID,
    },
  })

  const { data, loading, error } = useQuery(query, {
    variables: {
      id: messageID,
      number: contactMethodFromNumber,
    },
    fetchPolicy: 'network-only',
  })

  const status = data?.userContactMethod?.lastTestMessageState?.status ?? ''
  const contactMethodToNumber = data?.userContactMethod?.formattedValue ?? ''
  const contactMethodType = data?.userContactMethod?.type ?? ''
  const lastTestVerifyAt = data?.userContactMethod?.lastTestVerifyAt ?? ''
  const timeSinceLastVerified = now.diff(DateTime.fromISO(lastTestVerifyAt))
  contactMethodFromNumber = data?.phoneNumberInfo?.formatted ?? ''
  const errorMessage =
    (sendTestStatus?.error?.message ?? '') || (error?.message ?? '')

  useEffect(() => {
    if (data?.userContactMethod?.lastTestMessageState == null) {
      console.log('null data')
      return
    }
    if (loading) {
      console.log('loading')
      return
    }
    if (error) {
      console.log(contactMethodFromNumber)
      console.log('error: ', error)
      return
    }
    if (sendTestStatus.called) {
      console.log('already called mutation')
      return
    }
    // if (loading || error || sendTestStatus.called) {
    //   console.log("not calling mutation")
    //   return
    // }
    if (!(timeSinceLastVerified.as('seconds') < 60)) {
      console.log('mutation fired')
      sendTest()
    }
  }, [lastTestVerifyAt, loading])

  let details
  if (sendTestStatus.called && lastTestVerifyAt > now.toISO()) {
    details = data?.userContactMethod?.lastTestMessageState?.details ?? ''
  } else if (sendTestStatus.called) {
    details = 'Sending test message...'
  }

  const getTestStatusClass = (status: string): string => {
    switch (status) {
      case 'OK':
        return classes.statusOk
      case 'ERROR':
        return classes.statusError
      default:
        return classes.statusWarn
    }
  }

  if (loading || sendTestStatus.loading) return <Spinner text='Loading...' />

  return (
    <Dialog open onClose={onClose}>
      <DialogTitle>{title}</DialogTitle>

      <DialogContent>
        <DialogContentText>
          GoAlert is sending a {contactMethodType} to {contactMethodToNumber}{' '}
          from {contactMethodFromNumber}
        </DialogContentText>
        {details && (
          <DialogContentText className={getTestStatusClass(status)}>
            {toTitleCase(details)}
          </DialogContentText>
        )}
        {!details && (
          <DialogContentText className={classes.statusError}>
            Try again in one minute.
          </DialogContentText>
        )}
      </DialogContent>

      {errorMessage && (
        <DialogContentError>error = {errorMessage}</DialogContentError>
      )}

      <DialogActions>
        <Button color='primary' variant='contained' onClick={onClose}>
          Done
        </Button>
      </DialogActions>
    </Dialog>
  )
}

interface SendTestDialogProps {
  messageID: string
  onClose: (event: MouseEvent) => void
  disclaimer?: string
  title?: string
  subtitle?: string
}
