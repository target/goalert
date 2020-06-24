import React, { useState } from 'react'
import { Form } from '../forms'
import {
  Button,
  Card,
  CardActions,
  CardContent,
  Dialog,
  DialogActions,
  DialogTitle,
  InputAdornment,
  TextField,
} from '@material-ui/core'
import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'
import { useConfigValue } from '../util/RequireConfig'
import { AppLink } from '../util/AppLink'
import LoadingButton from '../loading/components/LoadingButton'
import DialogContentError from '../dialogs/components/DialogContentError'

const sendSMSMutation = gql`
  mutation DebugSendSMS($input: DebugSendSMSInput!) {
    debugSendSMS(input: $input) {
      id
      providerURL
    }
  }
`
/* TODO
  - Field errors
  - Style/padding/etc
  - Generic error display
*/
export default function AdminSMSSend(): JSX.Element {
  // const classes = useStyles()
  const [cfgFromNumber] = useConfigValue('Twilio.FromNumber')
  const [fromNumber, setFromNumber] = useState(cfgFromNumber)
  const [toNumber, setToNumber] = useState('')
  const [body, setBody] = useState('')
  const [showErrorDialog, setShowErrorDialog] = useState(false)

  const [send, sendStatus] = useMutation(sendSMSMutation, {
    variables: {
      input: {
        from: fromNumber,
        to: toNumber,
        body,
      },
    },
    onError: () => setShowErrorDialog(true),
  })

  return (
    <React.Fragment>
      <Form>
        <Card>
          <CardContent>
            <div>
              <TextField
                onChange={(e) => setFromNumber(e.target.value)}
                value={fromNumber}
                label='From Number'
                helperText='Please provide your country code e.g. +1 (USA), +91 (India), +44
              (UK)'
                type='tel'
                InputProps={{
                  startAdornment: (
                    <InputAdornment
                      position='start'
                      style={{ marginBottom: '0.1em' }}
                    >
                      +
                    </InputAdornment>
                  ),
                }}
              />
            </div>
            <br />
            <div>
              <TextField
                onChange={(e) => setToNumber(e.target.value)}
                value={toNumber}
                label='To Number'
                helperText='Please provide your country code e.g. +1 (USA), +91 (India), +44
              (UK)'
                type='tel'
                InputProps={{
                  startAdornment: (
                    <InputAdornment
                      position='start'
                      style={{ marginBottom: '0.1em' }}
                    >
                      +
                    </InputAdornment>
                  ),
                }}
              />
            </div>
            <br />
            <div>
              <TextField
                onChange={(e) => setBody(e.target.value)}
                value={body}
                fullWidth
                label='Body'
                InputLabelProps={{
                  shrink: true,
                }}
                multiline
              />
            </div>
          </CardContent>

          <CardActions>
            <LoadingButton
              buttonText='Send'
              onClick={() => {
                send()
              }}
              loading={sendStatus.loading}
            />
          </CardActions>

          <CardContent>
            {sendStatus.data?.debugSendSMS && (
              <AppLink to={sendStatus.data.debugSendSMS.providerURL} newTab>
                {sendStatus.data.debugSendSMS.id}
              </AppLink>
            )}
          </CardContent>
        </Card>
      </Form>

      <Dialog open={showErrorDialog} onClose={() => setShowErrorDialog(false)}>
        <DialogTitle>An error occurred</DialogTitle>
        <DialogContentError error={sendStatus.error?.message ?? ''} />
        <DialogActions>
          <Button
            color='primary'
            variant='contained'
            onClick={() => setShowErrorDialog(false)}
          >
            Okay
          </Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  )
}
