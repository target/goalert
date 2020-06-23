import React, { useState } from 'react'
import { Form } from '../forms'
import {
  Button,
  Card,
  CardContent,
  CardActions,
  Checkbox,
  Divider,
  List,
  ListItem,
  ListItemText,
  TextField,
  Tooltip,
  FormControlLabel,
} from '@material-ui/core'
import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'
import CopyText from '../util/CopyText'

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

  function renderListItem(label: string, text: string, copyText?: boolean) {
    return (
      <React.Fragment>
        <Divider />
        <ListItem>
          <ListItemText
            primary={label}
            secondary={
              copyText ? (
                <CopyText title={text} value={text} noUrl={true} />
              ) : (
                text
              )
            }
          />
        </ListItem>
      </React.Fragment>
    )
  }

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
          <Tooltip title='May incur Twilio charges' placement='right'>
            <Button
              onClick={() => {
                lookup()
              }}
              variant='contained'
              color='primary'
            >
              Lookup
            </Button>
          </Tooltip>
        </CardActions>

        {data?.debugPhoneNumberInfo && (
          <List dense={true}>
            {renderListItem(
              'Country Code',
              data.debugPhoneNumberInfo.countryCode,
            )}
            {renderListItem(
              'Region Code',
              data.debugPhoneNumberInfo.regionCode,
            )}
            {renderListItem(
              'Formatted Phone Number',
              data.debugPhoneNumberInfo.formatted,
            )}
            {data?.debugPhoneNumberInfo?.carrier && (
              <React.Fragment>
                {renderListItem(
                  'Carrier Name',
                  data.debugPhoneNumberInfo.carrier.name,
                  true,
                )}
                {renderListItem(
                  'Carrier Type',
                  data.debugPhoneNumberInfo.carrier.type,
                )}
                {renderListItem(
                  'Mobile Network Code',
                  data.debugPhoneNumberInfo.carrier.mobileNetworkCode,
                )}
                {renderListItem(
                  'Mobile Country Code',
                  data.debugPhoneNumberInfo.carrier.mobileCountryCode,
                )}
              </React.Fragment>
            )}
          </List>
        )}
      </Card>
    </Form>
  )
}
