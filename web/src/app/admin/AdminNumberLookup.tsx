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
  Grid,
  InputAdornment,
  List,
  ListItem,
  ListItemText,
  TextField,
  Tooltip,
  FormControlLabel,
} from '@material-ui/core'
import gql from 'graphql-tag'
import { useMutation, useQuery } from 'react-apollo'
import CopyText from '../util/CopyText'
import LoadingButton from '../loading/components/LoadingButton'
import DialogContentError from '../dialogs/components/DialogContentError'

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
    }
  }
`

export default function AdminNumberLookup(): JSX.Element {
  const [number, setNumber] = useState('')
  const [inclCarrier, setInclCarrier] = useState(false)
  const [showErrorDialog, setShowErrorDialog] = useState(false)

  const { data: numData, loading: numLoading, error: numError } = useQuery(
    numInfoQuery,
    {
      variables: { number: '+' + number },
      pollInterval: 0,
    },
  )

  const [lookup, { data, loading, error }] = useMutation(carrierInfoMut, {
    variables: { number: '+' + number },
    onError: () => setShowErrorDialog(true),
  })

  function renderListItem(label: string, text: string) {
    return (
      <React.Fragment>
        <Divider />
        <ListItem>
          <ListItemText
            primary={label}
            secondary={
              (text && <CopyText title={text} value={text} textOnly />) || ' '
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
                  onChange={(e) =>
                    setNumber(e.target.value.replace(/[^0-9]/g, ''))
                  }
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

          <List dense={true}>
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
            {(data?.debugCarrierInfo && (
              <React.Fragment>
                {renderListItem('Carrier Name', data.debugCarrierInfo.name)}
                {renderListItem('Carrier Type', data.debugCarrierInfo.type)}
                {renderListItem(
                  'Mobile Network Code',
                  data.debugCarrierInfo.mobileNetworkCode,
                )}
                {renderListItem(
                  'Mobile Country Code',
                  data.debugCarrierInfo.mobileCountryCode,
                )}
              </React.Fragment>
            )) || (
              <CardActions>
                <Tooltip title='May incur Twilio charges' placement='right'>
                  <LoadingButton
                    buttonText='Lookup Carrier Info'
                    onClick={() => {
                      lookup()
                    }}
                    loading={loading}
                  />
                </Tooltip>
              </CardActions>
            )}
          </List>
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
