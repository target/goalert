import React, { useState } from 'react'
import { Form } from '../forms'
import {
  Button,
  Card,
  CardContent,
  CardActions,
  Checkbox,
  Dialog,
  DialogActions,
  DialogTitle,
  Divider,
  InputAdornment,
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
import LoadingButton from '../loading/components/LoadingButton'
import DialogContentError from '../dialogs/components/DialogContentError'

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
  const [showErrorDialog, setShowErrorDialog] = useState(false)

  const [lookup, { data, loading, error }] = useMutation(
    inclCarrier ? mutation : mutNoCarrier,
    {
      variables: { number: '+' + number },
      onError: () => setShowErrorDialog(true),
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
    <React.Fragment>
      <Form>
        <Card>
          <CardContent>
            <TextField
              onChange={(e) => setNumber(e.target.value.replace(/^\+/, ''))}
              value={number}
              label='Phone Number'
              helperText='Please provide your country code e.g. +1 (USA)'
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
              <LoadingButton
                buttonText='Lookup'
                onClick={() => {
                  lookup()
                }}
                loading={loading}
              />
            </Tooltip>
          </CardActions>

          {data?.debugPhoneNumberInfo && (
            <List dense={true}>
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

      <Dialog open={showErrorDialog} onClose={() => setShowErrorDialog(false)}>
        <DialogTitle>An error occurred</DialogTitle>
        <DialogContentError error={error?.message ?? ''} />
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
