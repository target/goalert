import React from 'react'
import { gql, useQuery } from 'urql'
import { GenericError } from '../error-pages'

const checkAuthQuery = gql`
  {
    user {
      id
    }
  }
`

type RequireAuthProps = {
  children: React.ReactNode
  fallback: React.ReactNode
}
export default function RequireAuth(props: RequireAuthProps): React.ReactNode {
  const [{ error }] = useQuery({
    query: checkAuthQuery,
  })

  // if network/unauthorized display fallback
  if (error?.networkError) {
    return props.fallback
  }

  // if other error display generic
  if (error) {
    return <GenericError error={error.message} />
  }

  // otherwise display children
  return props.children
}
