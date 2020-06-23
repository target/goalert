import React, { useState } from 'react'
import { Form } from '../forms'
import {
  Card,
  CardContent,
  CardActions,
  Divider,
  List,
  ListItem,
  ListItemText,
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
        <CardContent>
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
        </CardContent>

        <CardActions>
          <Button
            onClick={() => {
              lookup()
            }}
            variant='contained'
            color='primary'
          >
            Lookup
          </Button>
        </CardActions>

        {data?.debugPhoneNumberInfo && (
          <React.Fragment>
            <Divider />
            <List dense={true}>
              <ListItem>
                <ListItemText
                  primary='Country Code'
                  secondary={data.debugPhoneNumberInfo.countryCode}
                />
              </ListItem>
              <Divider />
              <ListItem>
                <ListItemText
                  primary='Region Code'
                  secondary={data.debugPhoneNumberInfo.regionCode}
                />
              </ListItem>
              <Divider />
              <ListItem>
                <ListItemText
                  primary='Formatted Phone Number'
                  secondary={data.debugPhoneNumberInfo.formatted}
                />
              </ListItem>
              {data?.debugPhoneNumberInfo?.carrier && (
                <React.Fragment>
                  <Divider />
                  <ListItem>
                    <ListItemText
                      primary='Carrier Name'
                      secondary={data.debugPhoneNumberInfo.carrier.name}
                    />
                  </ListItem>
                  <Divider />
                  <ListItem>
                    <ListItemText
                      primary='Carrier Type'
                      secondary={data.debugPhoneNumberInfo.carrier.type}
                    />
                  </ListItem>
                  <Divider />
                  <ListItem>
                    <ListItemText
                      primary='Mobile Network Code'
                      secondary={
                        data.debugPhoneNumberInfo.carrier.mobileNetworkCode
                      }
                    />
                  </ListItem>
                  <Divider />
                  <ListItem>
                    <ListItemText
                      primary='Mobile Country Code'
                      secondary={
                        data.debugPhoneNumberInfo.carrier.mobileCountryCode
                      }
                    />
                  </ListItem>
                </React.Fragment>
              )}
            </List>
          </React.Fragment>
        )}
      </Card>
    </Form>
  )
}
