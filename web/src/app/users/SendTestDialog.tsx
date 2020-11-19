import React, { useEffect, useState, MouseEvent } from 'react'
import { useQuery, useMutation, gql } from '@apollo/client'

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
      id
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
  const errorMessage = error?.message ?? ''

  useEffect(() => {
    if (loading || error || sendTestStatus.called) {
      return
    }
    if (
      data?.userContactMethod?.lastTestMessageState == null ||
      !(timeSinceLastVerified.as('seconds') < 60)
    ) {
      sendTest()
    }
  }, [lastTestVerifyAt, loading])

  let details
  if (sendTestStatus.called && lastTestVerifyAt > now.toISO()) {
    details = data?.userContactMethod?.lastTestMessageState?.details ?? ''
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
          GoAlert is{' '}
          {contactMethodType === 'SMS'
            ? 'sending a test SMS message'
            : 'making a test voice call'}{' '}
          to {contactMethodToNumber} from {contactMethodFromNumber}.
        </DialogContentText>
        {details && (
          <DialogContentText className={getTestStatusClass(status)}>
            {toTitleCase(details)}
          </DialogContentText>
        )}
        {!details && (
          <DialogContentText className={classes.statusError}>
            Couldn't send a message yet, please try again after about a minute.
          </DialogContentText>
        )}
      </DialogContent>

      {errorMessage && <DialogContentError error={errorMessage} />}

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
