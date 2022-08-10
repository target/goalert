import React, { ReactNode } from 'react'
import { useSessionInfo } from '../../util/RequireConfig'
import { useResetURLParams, useURLParam } from '../../actions'
import { gql, useMutation } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'

const mutation = gql`
  mutation ($token: ID!) {
    linkAccountToken(token: $token)
  }
`

export default function AuthLink(): ReactNode {
  const [token] = useURLParam('authLinkToken', '')
  const [username] = useURLParam('username', '')
  const clearToken = useResetURLParams('authLinkToken')
  const clearUsername = useResetURLParams('username')
  const { ready } = useSessionInfo()

  const [linkAccount, linkAccountStatus] = useMutation(mutation, {
    variables: { token },
  })

  if (!token || !ready) {
    return null
  }

  return (
    <FormDialog
      title='Link Account?'
      confirm
      subTitle={`Click confirm to link this GoAlert account to slack user ${username}.`}
      errors={linkAccountStatus.error ? [linkAccountStatus.error] : []}
      onClose={() => {
        clearToken()
        clearUsername()
      }}
      onSubmit={() =>
        linkAccount().then(() => {
          clearToken()
          clearUsername()
        })
      }
    />
  )
}
