import React from 'react'
import AppLink from '../util/AppLink'
import { Target } from '../../schema'

export const UserLink = (user: Target): React.ReactNode => {
  return <AppLink to={`/users/${user.id}`}>{user.name}</AppLink>
}
