import React, { useState } from 'react'
import { Form } from '../forms'
import {
  Card,
  TextField,
  Button,
  Checkbox,
  FormControlLabel,
} from '@material-ui/core'
import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'

const mutNoCarrier = gql`
  mutation($number: String!) {
    debugPhoneNumberInfo(input: { number: $number }) {
      id
      countryCode
      regionCode
      formatted
    }
  }
`

const mutation = gql`
  mutation($number: String!) {
    debugPhoneNumberInfo(input: { number: $number }) {
      id
      countryCode
      regionCode
      formatted
      carrier {
        name
        type
        mobileNetworkCode
        mobileCountryCode
      }
    }
  }
`

/* TODO
  - Field errors
  - Style/padding/etc
  - Generic error display
  - Display data in readable way
*/

export default function AdminNumberLookup(): JSX.Element {
  const [number, setNumber] = useState('')
  const [inclCarrier, setInclCarrier] = useState(false)

  const [lookup, { data }] = useMutation(
    inclCarrier ? mutation : mutNoCarrier,
    {
      variables: { number },
    },
  )

  return (
    <Form>
      <Card>
        <TextField
          onChange={(e) => setNumber(e.target.value)}
          value={number}
          label='Phone Number'
          helperText='Including + and country code'
        />
        <FormControlLabel
          control={
            <Checkbox
              checked={inclCarrier}
              onChange={(e) => setInclCarrier(e.target.checked)}
            />
          }
          label='Include carrier information'
        />

        <Button
          onClick={() => {
            lookup()
          }}
        >
          Lookup
        </Button>
        {data?.debugPhoneNumberInfo && (
          <div>
            <hr />
            {JSON.stringify(data, null, '  ')}
          </div>
        )}
      </Card>
    </Form>
  )
}
