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
import { DateTime, Duration } from 'luxon'
import {
  DestinationInput,
  NotificationState,
  UserContactMethod,
} from '../../schema'
import DestinationInputChip from '../util/DestinationInputChip'
import { Time } from '../util/Time'

const query = gql`
  query ($id: ID!) {
    userContactMethod(id: $id) {
      id
      dest {
        type
        args
      }
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

function getTestStatusColor(status?: string | null): string {
  switch (status) {
    case 'OK':
      return 'success'
    case 'ERROR':
      return 'error'
    default:
      return 'warning'
  }
}

type SendTestContentProps = {
  dest: DestinationInput
  isSending: boolean
  isWaiting: boolean

  sentTime?: string | null
  sentState?: NotificationState | null
}

export function SendTestContent(props: SendTestContentProps): React.ReactNode {
  return (
    <React.Fragment>
      <DialogContent>
        <DialogContentText>
          GoAlert is sending a test to{' '}
          <DestinationInputChip value={props.dest} />.
        </DialogContentText>

        {props.sentTime && (
          <DialogContentText style={{ marginTop: '1rem' }}>
            The test message was scheduled for delivery at{' '}
            <Time time={props.sentTime} />.
          </DialogContentText>
        )}
        {props.sentState?.formattedSrcValue && (
          <DialogContentText>
            <b>Sender:</b> {props.sentState.formattedSrcValue}
          </DialogContentText>
        )}

        {props.isWaiting && <Spinner text='Waiting to send (< 1 min)...' />}
        {props.isSending && <Spinner text='Sending Test...' />}
        {props.sentState?.details === 'Pending' ? (
          <Spinner text='Waiting in queue...' />
        ) : (
          <DialogContentText
            color={getTestStatusColor(props.sentState?.status)}
          >
            <b>Status:</b> {toTitleCase(props.sentState?.details || '')}
          </DialogContentText>
        )}
      </DialogContent>
    </React.Fragment>
  )
}

export default function SendTestDialog(
  props: SendTestDialogProps,
): JSX.Element {
  const { onClose, contactMethodID } = props
  const [sendTestStatus, sendTest] = useMutation(mutation)

  const [cmInfo, refreshCMInfo] = useQuery<{
    userContactMethod: UserContactMethod
  }>({
    query,
    variables: {
      id: contactMethodID,
    },
    requestPolicy: 'network-only',
  })

  // Should not happen, but just in case.
  if (cmInfo.error) throw cmInfo.error
  const cm = cmInfo.data?.userContactMethod
  if (!cm) throw new Error('missing contact method') // should be impossible (since we already checked the error)

  // We expect the status to update over time, so we manually refresh
  // as long as the dialog is open.
  useEffect(() => {
    const t = setInterval(refreshCMInfo, 3000)
    return () => clearInterval(t)
  }, [])

  // We keep a stable timestamp to track how long the dialog has been open.
  const [now] = useState(DateTime.utc())

  const isWaitingToSend =
    (cm.lastTestVerifyAt
      ? now.diff(DateTime.fromISO(cm.lastTestVerifyAt))
      : Duration.fromObject({ day: 1 })
    ).as('seconds') < 60

  // already sent a test message recently
  const [alreadySent, setAlreadySent] = useState(
    !!cm.lastTestMessageState && isWaitingToSend,
  )

  useEffect(() => {
    if (alreadySent) return

    // wait until at least a minute has passed
    if (isWaitingToSend) return

    sendTest(
      { id: contactMethodID },
      {
        additionalTypenames: ['UserContactMethod'],
      },
    )

    setAlreadySent(true)
  }, [isWaitingToSend, alreadySent])

  return (
    <Dialog open onClose={onClose}>
      <DialogTitle>Test Contact Method</DialogTitle>
      <SendTestContent
        dest={cm.dest}
        isWaiting={isWaitingToSend && !alreadySent}
        isSending={sendTestStatus.fetching}
        sentState={
          cm.lastTestMessageState && alreadySent
            ? cm.lastTestMessageState
            : undefined
        }
        sentTime={
          cm.lastTestMessageState && alreadySent ? cm.lastTestVerifyAt : null
        }
      />
      <DialogActions>
        <Button color='primary' variant='contained' onClick={onClose}>
          Done
        </Button>
      </DialogActions>
    </Dialog>
  )
}

interface SendTestDialogProps {
  contactMethodID: string
  onClose: (event: MouseEvent) => void
}
