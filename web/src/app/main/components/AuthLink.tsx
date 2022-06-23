import React from 'react'
import { useSessionInfo } from '../../util/RequireConfig'
import { useResetURLParams, useURLParam } from '../../actions'
import { gql, useMutation } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'

const mutation = gql`
  mutation ($token: ID!) {
    linkAccountToken(token: $token)
  }
`

export default function AuthLink() {
  const [token] = useURLParam('authLinkToken', '')
  const clearToken = useResetURLParams('authLinkToken')
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
      subTitle='Click confirm to link this GoAlert account.'
      errors={linkAccountStatus.error ? [linkAccountStatus.error] : []}
      onClose={() => clearToken()}
      onSubmit={() =>
        linkAccount().then(() => {
          clearToken()
        })
      }
    />
  )
}
