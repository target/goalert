import React from 'react'
import { AppLink } from '../util/AppLink'

export const UserLink = user => {
  return <AppLink to={`/users/${user.id}`}>{user.name}</AppLink>
}
