import React, { useState } from 'react'
import { gql, useQuery, useMutation } from 'urql'
import { Form } from '../forms'
import {
  Button,
  Card,
  CardActions,
  CardContent,
  Dialog,
  DialogActions,
  DialogTitle,
  Grid,
  TextField,
  Typography,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'

import { useConfigValue } from '../util/RequireConfig'
import AppLink from '../util/AppLink'
import TelTextField from '../util/TelTextField'
import LoadingButton from '../loading/components/LoadingButton'
import DialogContentError from '../dialogs/components/DialogContentError'
import FromValueField from '../util/FromValueField'

const debugMessageStatusQuery = gql`
  query DebugMessageStatus($input: DebugMessageStatusInput!) {
    debugMessageStatus(input: $input) {
      state {
        details
        status
        formattedSrcValue
      }
    }
  }
`
const sendSMSMutation = gql`
  mutation DebugSendSMS($input: DebugSendSMSInput!) {
    debugSendSMS(input: $input) {
      id
      providerURL
      fromNumber
    }
  }
`

const useStyles = makeStyles({
  twilioLink: {
    display: 'flex',
    alignItems: 'center',
  },
})

export default function AdminSMSSend(): React.ReactNode {
  const classes = useStyles()
  const [cfgFromNumber, cfgSID] = useConfigValue(
    'Twilio.FromNumber',
    'Twilio.MessagingServiceSID',
  ) as [string | null, string | null]
  const [messageID, setMessageID] = useState('')
  const [fromNumber, setFromNumber] = useState(cfgSID || cfgFromNumber || '')
  const [toNumber, setToNumber] = useState('')
  const [body, setBody] = useState('')
  const [showErrorDialog, setShowErrorDialog] = useState(false)

  const [{ data: smsData, fetching: smsLoading, error: smsError }, commit] =
    useMutation(sendSMSMutation)

  const [{ data }] = useQuery({
    query: debugMessageStatusQuery,
    variables: {
      input: { providerMessageID: messageID },
    },
    pause: !messageID,
  })

  const isSent = data?.debugMessageStatus?.state?.status === 'OK'
  let _details = data?.debugMessageStatus?.state?.details || 'Sending...'
  _details = _details.charAt(0).toUpperCase() + _details.slice(1)

  const details = isSent
    ? `${_details} from ${data?.debugMessageStatus?.state?.formattedSrcValue}.`
    : _details

  return (
    <React.Fragment>
      <Form
        onSubmit={(e: { preventDefault: () => void }) => {
          e.preventDefault()
          commit({
            input: {
              from: fromNumber,
              to: toNumber,
              body,
            },
          }).then((res) => {
            if (res.error) {
              setShowErrorDialog(true)
              return
            }
            setMessageID(res.data.debugSendSMS.id)
          })
        }}
      >
        <Card>
          <CardContent>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={12} md={12} lg={6}>
                <FromValueField
                  onChange={(e) => setFromNumber(e.target.value)}
                  value={fromNumber}
                  defaultPhone={cfgFromNumber}
                  defaultSID={cfgSID}
                  fullWidth
                />
              </Grid>
              <Grid item xs={12} sm={12} md={12} lg={6}>
                <TelTextField
                  onChange={(e) => setToNumber(e.target.value)}
                  value={toNumber}
                  fullWidth
                  label='To Number'
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  onChange={(e) => setBody(e.target.value)}
                  value={body}
                  fullWidth
                  label='Body'
                  multiline
                />
              </Grid>
            </Grid>
          </CardContent>

          <CardActions>
            <LoadingButton buttonText='Send' loading={smsLoading} />
            {smsData?.debugSendSMS && (
              <AppLink to={smsData.debugSendSMS.providerURL} newTab>
                <div className={classes.twilioLink}>
                  <Typography>{details} Open in Twilio&nbsp;</Typography>
                  <OpenInNewIcon fontSize='small' />
                </div>
              </AppLink>
            )}
          </CardActions>
        </Card>
      </Form>

      <Dialog open={showErrorDialog} onClose={() => setShowErrorDialog(false)}>
        <DialogTitle>An error occurred</DialogTitle>
        <DialogContentError error={smsError?.message ?? ''} />
        <DialogActions>
          <Button variant='contained' onClick={() => setShowErrorDialog(false)}>
            Okay
          </Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  )
}
