import React, { useState } from 'react'
import { Form } from '../forms'
import { Card, TextField, Button } from '@material-ui/core'
import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'

const query = gql`
  query($number: String!) {
    phoneNumberInfo(number: $number) {
      id
      formatted
    }
  }
`

const queryCarrier = gql`
  query($number: String!) {
    phoneNumberInfo(number: $number) {
      id
      formatted
    }
  }
`

export default function AdminNumberLookup(): JSX.Element {
  const [number, setNumber] = useState('')
  const [inclCarrier, setInclCarrier] = useState(false)
  const [submit, setSubmit] = useState(false)

  const { data, loading, error } = useQuery(
    inclCarrier ? queryCarrier : query,
    {
      variables: { number },
      pollInterval: 0,
      skip: !submit,
    },
  )

  return (
    <Form>
      <Card>
        <TextField
          onChange={(e) => {
            setSubmit(false)
            setNumber(e.target.value)
          }}
          value={number}
          label='Phone Number'
          helperText='Including + and country code'
        />

        <Button onClick={() => setSubmit(true)}>Lookup</Button>
        {data?.phoneNumberInfo && (
          <div>
            <hr />
            {JSON.stringify(data, null, '  ')}
          </div>
        )}
      </Card>
    </Form>
  )
}
