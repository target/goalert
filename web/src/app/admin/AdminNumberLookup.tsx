import React, { useState, useEffect } from 'react'
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
  List,
  ListItem,
  ListItemText,
  Tooltip,
} from '@mui/material'
import { useMutation, useQuery, gql, CombinedError } from 'urql'
import CopyText from '../util/CopyText'
import TelTextField from '../util/TelTextField'
import LoadingButton from '../loading/components/LoadingButton'
import DialogContentError from '../dialogs/components/DialogContentError'

import { PhoneNumberInfo, DebugCarrierInfo } from '../../schema'

const carrierInfoMut = gql`
  mutation ($number: String!) {
    debugCarrierInfo(input: { number: $number }) {
      name
      type
      mobileNetworkCode
      mobileCountryCode
    }
  }
`

const numInfoQuery = gql`
  query ($number: String!) {
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

  const [{ data: numData, error: queryError }] = useQuery({
    query: numInfoQuery,
    variables: { number },
  })
  const numInfo = numData?.phoneNumberInfo as PhoneNumberInfo

  const [
    { data: carrData, fetching: carrLoading, error: mutationError },
    commit,
  ] = useMutation(carrierInfoMut)
  const carrInfo = carrData?.debugCarrierInfo as DebugCarrierInfo

  const [lastError, setLastError] = useState<CombinedError | null>(null)
  useEffect(() => {
    if (queryError) setLastError(queryError)
  }, [queryError])
  useEffect(() => {
    if (mutationError) setLastError(mutationError)
  }, [mutationError])

  function renderListItem(label: string, text = ''): JSX.Element {
    return (
      <React.Fragment>
        <Divider />
        <ListItem>
          <ListItemText
            primary={label}
            secondary={(text && <CopyText title={text} value={text} />) || '?'}
          />
        </ListItem>
      </React.Fragment>
    )
  }

  return (
    <React.Fragment>
      <Form
        onSubmit={(e: { preventDefault: () => void }) => {
          e.preventDefault()
          commit({ number })
          setStaleCarrier(false)
        }}
      >
        <Card>
          <CardContent>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TelTextField
                  onChange={(e) => {
                    setNumber(e.target.value)
                    setStaleCarrier(true)
                  }}
                  value={number}
                  label='Phone Number'
                />
              </Grid>
            </Grid>
          </CardContent>

          <List dense>
            {renderListItem('Country Code', numInfo?.countryCode)}
            {renderListItem('Formatted Phone Number', numInfo?.formatted)}
            {renderListItem('Region Code', numInfo?.regionCode)}
            {renderListItem(
              'Valid',
              numInfo?.valid
                ? 'true'
                : `false` + (numInfo?.error ? ` (${numInfo?.error})` : ''),
            )}
            {(carrInfo && !staleCarrier && !carrLoading && (
              <React.Fragment>
                {renderListItem('Carrier Name', carrInfo.name)}
                {renderListItem('Carrier Type', carrInfo.type)}
                {renderListItem(
                  'Mobile Network Code',
                  carrInfo.mobileNetworkCode,
                )}
                {renderListItem(
                  'Mobile Country Code',
                  carrInfo.mobileCountryCode,
                )}
              </React.Fragment>
            )) || (
              <CardActions>
                <Tooltip title='May incur Twilio charges' placement='right'>
                  <LoadingButton
                    buttonText='Lookup Carrier Info'
                    disabled={!numInfo?.valid}
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
          <Button variant='contained' onClick={() => setLastError(null)}>
            Okay
          </Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  )
}
