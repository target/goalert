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
    username: '',
    alertID: '',
  })
  const [, navigate] = useLocation()

  const resetParams = useResetURLParams('authLinkToken', 'username')
  const { ready } = useSessionInfo()

  const [linkAccount, linkAccountStatus] = useMutation(mutation, {
    variables: { token: params.authLinkToken },
  })

  if (!params.username || !params.authLinkToken || !ready) {
    return null
  }

  return (
    <FormDialog
      title='Link Account?'
      confirm
      subTitle={`Click confirm to link this GoAlert account to slack user @${params.username}. You will be able to update alerts from slack when your account has been linked.`}
      errors={linkAccountStatus.error ? [linkAccountStatus.error] : []}
      onClose={() => {
        resetParams()
      }}
      onSubmit={() =>
        linkAccount().then(() => {
          navigate(`/alerts/${params.alertID}`)
          resetParams()
        })
      }
    />
  )
}
