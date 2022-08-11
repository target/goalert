import React from 'react'
import { useSessionInfo } from '../../util/RequireConfig'
import { useResetURLParams, useURLParams } from '../../actions'
import { gql, useMutation } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import { useLocation } from 'wouter'
import Snackbar from '@mui/material/Snackbar'
import { Alert } from '@mui/material'

const mutation = gql`
  mutation ($token: ID!) {
    linkAccountToken(token: $token)
  }
`

const updateStatusMutation = gql`
  mutation UpdateAlertsMutation($input: UpdateAlertsInput!) {
    updateAlerts(input: $input) {
      id
    }
  }
`

export default function AuthLink(): JSX.Element {
  const [params] = useURLParams({
    authLinkToken: '',
    details: '',
    alertID: '',
    action: '',
  })
  const [, navigate] = useLocation()

  const resetParams = useResetURLParams('authLinkToken', 'details')
  const { ready } = useSessionInfo()

  const [linkAccountStatus, linkAccount] = useMutation(mutation)
  const [, updateAlertStatus] = useMutation(updateStatusMutation)

  if (!params.details || !params.authLinkToken || !ready) {
    return <div>{undefined}</div>
  }

  const authTokenExpired = linkAccountStatus?.error?.message.includes('expired')

  if (linkAccountStatus.error) {
    return (
      <Snackbar
        anchorOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
        autoHideDuration={6000}
        onClose={() => console.log('hello')}
        open={authTokenExpired || Boolean(linkAccountStatus?.error)}
      >
        <Alert severity='error'>
          {authTokenExpired
            ? 'The auth link token has expired. Please try again later.'
            : 'An unexpected error has occurred. Please try again later.'}
        </Alert>
      </Snackbar>
    )
  }

  return (
    <FormDialog
      title='Link Account?'
      confirm
      subTitle={`Click confirm to link this account to ${params.details}.`}
      errors={linkAccountStatus.error ? [linkAccountStatus.error] : []}
      onClose={() => {
        resetParams()
      }}
      onSubmit={() =>
        linkAccount({ token: params.authLinkToken }).then((result) => {
          if (result.error) return
          if (params.alertID) navigate(`/alerts/${params.alertID}`)
          if (params.action) {
            updateAlertStatus({
              input: {
                alertIDs: [params.alertID],
                newStatus:
                  params.action === 'ResultAcknowledge'
                    ? 'StatusAcknowledged'
                    : 'StatusClosed',
              },
            })
          }

          resetParams()
        })
      }
    />
  )
}
