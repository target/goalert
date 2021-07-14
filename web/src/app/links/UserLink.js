import React from 'react'
import { MuiLink } from '../util/AppLink'

export const UserLink = (user) => {
  return <MuiLink to={`/users/${user.id}`}>{user.name}</MuiLink>
}
