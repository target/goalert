import React, { useState } from 'react'
import { Form } from '../forms'
import {
  Button,
  Card,
  CardContent,
  CardActions,
  Dialog,
  DialogActions,
  DialogTitle,
  Divider,
  Grid,
  InputAdornment,
  List,
  ListItem,
  ListItemText,
  TextField,
  Tooltip,
} from '@material-ui/core'
import gql from 'graphql-tag'
import { useMutation, useQuery } from 'react-apollo'
import CopyText from '../util/CopyText'
import LoadingButton from '../loading/components/LoadingButton'
import DialogContentError from '../dialogs/components/DialogContentError'
import { ApolloError } from 'apollo-client'

const carrierInfoMut = gql`
  mutation($number: String!) {
    debugCarrierInfo(input: { number: $number }) {
      name
      type
      mobileNetworkCode
      mobileCountryCode
    }
  }
`

const numInfoQuery = gql`
  query($number: String!) {
    phoneNumberInfo(number: $number) {
      id
      valid
      regionCode
      countryCode
      formatted
      error
    }
  }
`

export default function AdminNumberLookup(): JSX.Element {
  const [number, setNumber] = useState('')
  const [staleCarrier, setStaleCarrier] = useState(true)
  const [lastError, setLastError] = useState(null as null | ApolloError)

  const { data: numData } = useQuery(numInfoQuery, {
    variables: { number: '+' + number },
    pollInterval: 0,
    onError: (err) => setLastError(err),
  })

  const [lookup, { data: carrData, loading: carrLoading }] = useMutation(
    carrierInfoMut,
    {
      variables: { number: '+' + number },
      onError: (err) => setLastError(err),
    },
  )

  function renderListItem(label: string, _text: string): JSX.Element {
    const text = (_text === undefined ? '' : _text).toString()
    return (
      <React.Fragment>
        <Divider />
        <ListItem>
          <ListItemText
            primary={label}
            secondary={
              (text && <CopyText title={text} value={text} textOnly />) || '?'
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
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  onChange={(e) => {
                    setNumber(e.target.value.replace(/[^0-9]/g, ''))
                    setStaleCarrier(true)
                  }}
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
              </Grid>
            </Grid>
          </CardContent>

          <List dense>
            {renderListItem(
              'Country Code',
              numData?.phoneNumberInfo?.countryCode,
            )}
            {renderListItem(
              'Formatted Phone Number',
              numData?.phoneNumberInfo?.formatted,
            )}
            {renderListItem(
              'Region Code',
              numData?.phoneNumberInfo?.regionCode,
            )}
            {renderListItem(
              'Valid',
              numData?.phoneNumberInfo?.valid
                ? 'true'
                : `false` +
                    (numData?.phoneNumberInfo?.error
                      ? ` (${numData?.phoneNumberInfo?.error})`
                      : ''),
            )}
            {(carrData?.debugCarrierInfo && !staleCarrier && !carrLoading && (
              <React.Fragment>
                {renderListItem('Carrier Name', carrData.debugCarrierInfo.name)}
                {renderListItem('Carrier Type', carrData.debugCarrierInfo.type)}
                {renderListItem(
                  'Mobile Network Code',
                  carrData.debugCarrierInfo.mobileNetworkCode,
                )}
                {renderListItem(
                  'Mobile Country Code',
                  carrData.debugCarrierInfo.mobileCountryCode,
                )}
              </React.Fragment>
            )) || (
              <CardActions>
                <Tooltip title='May incur Twilio charges' placement='right'>
                  <LoadingButton
                    buttonText='Lookup Carrier Info'
                    onClick={() => {
                      lookup()
                      setStaleCarrier(false)
                    }}
                    disabled={!numData?.phoneNumberInfo?.valid}
                    loading={carrLoading}
                  />
                </Tooltip>
              </CardActions>
            )}
          </List>
        </Card>
      </Form>

      <Dialog open={Boolean(lastError)} onClose={() => setLastError(null)}>
        <DialogTitle>An error occurred</DialogTitle>
        <DialogContentError error={lastError?.message ?? ''} />
        <DialogActions>
          <Button
            color='primary'
            variant='contained'
            onClick={() => setLastError(null)}
          >
            Okay
          </Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  )
}
