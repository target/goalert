import React, { useEffect, MouseEvent, useState } from 'react'
import { gql, useQuery, useMutation } from 'urql'

import Spinner from '../loading/components/Spinner'

import {
  Button,
  Dialog,
  DialogActions,
  DialogTitle,
  DialogContent,
  DialogContentText,
} from '@mui/material'
import toTitleCase from '../util/toTitleCase'
import DialogContentError from '../dialogs/components/DialogContentError'
import { DateTime } from 'luxon'
import { ContactMethodType } from '../../schema'

const query = gql`
  query ($id: ID!) {
    userContactMethod(id: $id) {
      id
      type
      formattedValue
      lastTestVerifyAt
      lastTestMessageState {
        details
        status
        formattedSrcValue
      }
    }
  }
`

const mutation = gql`
  mutation ($id: ID!) {
    testContactMethod(id: $id)
  }
`

export default function SendTestDialog(
  props: SendTestDialogProps,
): JSX.Element {
  const { title = 'Test Delivery Status', onClose, messageID } = props

  const [sendTestStatus, sendTest] = useMutation(mutation)

  const [{ data, fetching, error }] = useQuery({
    query,
    variables: {
      id: messageID,
    },
    requestPolicy: 'network-only',
  })

  // We keep a stable timestampe to track how long the dialog has been open
  const [now] = useState(DateTime.utc())
  const status = data?.userContactMethod?.lastTestMessageState?.status ?? ''
  const cmDestValue = data?.userContactMethod?.formattedValue ?? ''
  const cmType: ContactMethodType = data?.userContactMethod?.type ?? ''
  const lastTestVerifyAt = data?.userContactMethod?.lastTestVerifyAt ?? ''
  const timeSinceLastVerified = now.diff(DateTime.fromISO(lastTestVerifyAt))
  const fromValue =
    data?.userContactMethod?.lastTestMessageState?.formattedSrcValue ?? ''
  const errorMessage = (error?.message || sendTestStatus.error?.message) ?? ''

  useEffect(() => {
    if (fetching || errorMessage || sendTestStatus.data) {
      return
    }
    if (
      data?.lastTestMessageState == null ||
      !(timeSinceLastVerified.as('seconds') < 60)
    ) {
      sendTest({ id: messageID })
    }
  }, [lastTestVerifyAt, fetching])

  let details = ''
  if (sendTestStatus.data && lastTestVerifyAt > now.toISO()) {
    details = data?.userContactMethod?.lastTestMessageState?.details ?? ''
  }

  const isLoading =
    sendTestStatus.fetching ||
    (!!details && !!errorMessage) ||
    status === 'pending'

  const getTestStatusColor = (status: string): string => {
    switch (status) {
      case 'OK':
        return 'success'
      case 'ERROR':
        return 'error'
      default:
        return 'warning'
    }
  }

  const msg = (): string => {
    switch (cmType) {
      case 'SMS':
      case 'VOICE':
        return `${
          cmType === 'SMS' ? 'SMS message' : 'voice call'
        } to ${cmDestValue}`
      case 'EMAIL':
        return `email to ${cmDestValue}`
      default:
        return `to ${cmDestValue}`
    }
  }

  return (
    <Dialog open onClose={onClose}>
      <DialogTitle>{title}</DialogTitle>

      <DialogContent>
        <DialogContentText>
          GoAlert is sending a test {msg()}.
        </DialogContentText>
        {isLoading && <Spinner text='Sending Test...' />}
        {fromValue && (
          <DialogContentText>
            The test message was sent from {fromValue}.
          </DialogContentText>
        )}
        {!!details && (
          <DialogContentText color={getTestStatusColor(status)}>
            {toTitleCase(details)}
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
