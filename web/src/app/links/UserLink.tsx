import React from 'react'
import AppLink from '../util/AppLink'
import { Target } from '../../schema'

export const UserLink = (user: Target): JSX.Element => {
  return <AppLink to={`/users/${user.id}`}>{user.name}</AppLink>
}
