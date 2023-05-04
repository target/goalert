import React, { useEffect, useState } from 'react'
import { useSessionInfo } from '../../util/RequireConfig'
import { useURLParam } from '../../actions'
import { gql, useMutation, useQuery } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import { useLocation } from 'wouter'
import Snackbar from '@mui/material/Snackbar'
import { Alert, Grid, Typography } from '@mui/material'
import { LinkAccountInfo } from '../../../schema'

const mutation = gql`
  mutation ($token: ID!) {
    linkAccount(token: $token)
  }
`
const query = gql`
  query ($token: ID!) {
    linkAccountInfo(token: $token) {
      userDetails
      alertID
      alertNewStatus
    }
  }
`

const updateStatusMutation = gql`
  mutation UpdateAlertsMutation($input: UpdateAlertsInput!) {
    updateAlerts(input: $input) {
      id
    }
  }
`

export default function AuthLink(): JSX.Element | null {
  const [token, setToken] = useURLParam('authLinkToken', '')
  const [, navigate] = useLocation()

  const { ready, userName } = useSessionInfo()

  const [{ data, fetching, error }] = useQuery({
    query,
    variables: { token },
    pause: !token,
  })
  const [linkAccountStatus, linkAccount] = useMutation(mutation)
  const [, updateAlertStatus] = useMutation(updateStatusMutation)
  const [snack, setSnack] = useState(true)

  const info: LinkAccountInfo = data?.linkAccountInfo

  useEffect(() => {
    if (!ready) return
    if (!token) return
    if (fetching) return
    if (error) return
    if (info) return

    setToken('')
  }, [!!info, !!error, fetching, ready, token])

  if (!token || !ready || fetching) {
    return null
  }

  if (error) {
    return (
      <Snackbar
        anchorOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
        autoHideDuration={6000}
        onClose={() => setSnack(false)}
        open={snack && !!error}
      >
        <Alert severity='error'>
          Unable to fetch account link details. Try again later.
        </Alert>
      </Snackbar>
    )
  }

  if (!info) {
    return (
      <Snackbar
        anchorOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
        autoHideDuration={6000}
        onClose={() => setSnack(false)}
        open={snack}
      >
        <Alert severity='error'>
          Invalid or expired account link URL. Try again.
        </Alert>
      </Snackbar>
    )
  }

  let alertAction = ''
  if (info.alertID && info.alertNewStatus) {
    switch (info.alertNewStatus) {
      case 'StatusAcknowledged':
        alertAction = `alert #${data.linkAccountInfo.alertID} will be acknowledged.`
        break
      case 'StatusClosed':
        alertAction = `alert #${data.linkAccountInfo.alertID} will be closed.`
        break
      default:
        alertAction = `Alert #${data.linkAccountInfo.alertID} will be updated to ${info.alertNewStatus}.`
        break
    }
  }

  return (
    <FormDialog
      title='Link Account?'
      confirm
      errors={linkAccountStatus.error ? [linkAccountStatus.error] : []}
      onClose={() => setToken('')}
      onSubmit={() =>
        linkAccount({ token }).then((result) => {
          if (result.error) return
          if (info.alertID) navigate(`/alerts/${info.alertID}`)
          if (info.alertNewStatus) {
            updateAlertStatus({
              input: {
                alertIDs: [info.alertID],
                newStatus: info.alertNewStatus,
              },
            })
          }

          setToken('')
        })
      }
      form={
        <Grid container spacing={2}>
          <Grid item xs={12}>
            <Typography>
              Clicking confirm will link the current GoAlert user{' '}
              <b>{userName}</b> with:
            </Typography>
          </Grid>
          <Grid item xs={12}>
            <Typography>{data.linkAccountInfo.userDetails}.</Typography>
          </Grid>
          {alertAction && (
            <Grid item xs={12}>
              <Typography>After linking, {alertAction}</Typography>
            </Grid>
          )}
        </Grid>
      }
    />
  )
}
