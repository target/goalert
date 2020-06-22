import React, { useState } from 'react'
import { Form } from '../forms'
import { Card, TextField, Button } from '@material-ui/core'
import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'
import { useConfigValue } from '../util/RequireConfig'
import Spinner from '../loading/components/Spinner'
import { AppLink } from '../util/AppLink'

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
  const [cfgFromNumber] = useConfigValue('Twilio.FromNumber')
  const [fromNumber, setFromNumber] = useState(cfgFromNumber)
  const [toNumber, setToNumber] = useState('')
  const [body, setBody] = useState('')

  const [send, sendStatus] = useMutation(sendSMSMutation, {
    variables: {
      input: {
        from: fromNumber,
        to: toNumber,
        body,
      },
    },
  })

  return (
    <Form>
      <Card>
        <TextField
          onChange={(e) => setFromNumber(e.target.value)}
          fullWidth
          value={fromNumber}
          label='From Number'
          helperText='Including + and country code'
        />
        <TextField
          onChange={(e) => setToNumber(e.target.value)}
          fullWidth
          value={toNumber}
          label='To Number'
          helperText='Including + and country code'
        />
        <TextField
          onChange={(e) => setBody(e.target.value)}
          fullWidth
          value={body}
          label='Body'
          multiline
        />

        <Button
          onClick={() => {
            send()
          }}
        >
          Send
        </Button>

        <hr />
        {sendStatus.loading && <Spinner />}
        {sendStatus.error && sendStatus.error.message}
        {sendStatus.data?.debugSendSMS && (
          <AppLink to={sendStatus.data.debugSendSMS.providerURL} newTab>
            {sendStatus.data.debugSendSMS.id}
          </AppLink>
        )}
      </Card>
    </Form>
  )
}
