import React, { useEffect, useState } from 'react'
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

import moment from 'moment'
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

export default function SendTestDialog(props) {
  const classes = useStyles()

  const { title = 'Test Delivery Status', onClose, messageID } = props

  let [contactMethodFromNumber] = useConfigValue('Twilio.FromNumber')

  const [mutationSent, setMutationSent] = useState(false)

  const [sendTest, sendTestStatus] = useMutation(mutation, {
    variables: {
      id: messageID,
    },
  })

  const { data, loading, error, refetch } = useQuery(query, {
    variables: {
      id: messageID,
      number: contactMethodFromNumber,
    },
    skip:
      sendTestStatus.error || sendTestStatus.loading || mutationSent === false,
  })

  const currentTimeMinusOneMinute = DateTime.local()
    .minus({ minutes: 1 })
    .toISO()

  const details = data?.userContactMethod?.lastTestMessageState?.details ?? ''
  const status = data?.userContactMethod?.lastTestMessageState?.status ?? ''

  const contactMethodToNumber = data?.userContactMethod?.formattedValue ?? ''
  const contactMethodType = data?.userContactMethod?.type ?? ''
  const lastTestVerifyAt = data?.userContactMethod?.lastTestVerifyAt ?? ''
  contactMethodFromNumber = data?.phoneNumberInfo?.formatted ?? ''
  const errorMessage =
    (sendTestStatus?.error?.message ?? '') || (error?.message ?? '')

  useEffect(() => {
    // if mutation is not sent and the last verified time was over one minute, send the mutation
    if (!mutationSent && lastTestVerifyAt < currentTimeMinusOneMinute) {
      console.log('not sent, verified over a minute ago, send test')
      setMutationSent(true)
      sendTest()
      refetch()
    }
    // if there is a mutation error and the last verified time was over one minute ago, allow retry
    if (errorMessage && lastTestVerifyAt < currentTimeMinusOneMinute) {
      console.log(
        'send test error, verified over a minute ago, attempt send test',
      )
      sendTest()
      refetch()
    }
    // if there is data and the last verified time was less than one minute ago, display the details
    if (
      details !== null &&
      (lastTestVerifyAt > currentTimeMinusOneMinute ||
        sendTestStatus.error === null)
    ) {
      console.log('sent less than a minute ago, display last details')
      setMutationSent(false)
    }
  }, [])

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
