import React, { ReactNode } from 'react'
import { useSessionInfo } from '../../util/RequireConfig'
import { useResetURLParams, useURLParams } from '../../actions'
import { gql, useMutation } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'
import { useLocation } from 'wouter'

const mutation = gql`
  mutation ($token: ID!) {
    linkAccountToken(token: $token)
  }
`

export default function AuthLink(): ReactNode {
  const [params] = useURLParams({
    authLinkToken: '',
    details: '',
    alertID: '',
    action: '',
  })
  const [, navigate] = useLocation()

  const resetParams = useResetURLParams('authLinkToken', 'details')
  const { ready } = useSessionInfo()

  const [linkAccount, linkAccountStatus] = useMutation(mutation, {
    variables: { token: params.authLinkToken },
  })

  if (!params.details || !params.authLinkToken || !ready) {
    return null
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
        linkAccount().then(() => {
          if (params.alertID) {
            navigate(`/alerts/${params.alertID}`)
          }
          if (params.action) {
            // make request to close/ack here
            // if fail trigger toast
          }

          // always call
          resetParams()
        })
      }
    />
  )
}
