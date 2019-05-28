import React from 'react'
import { Link } from 'react-router-dom'

export const UserLink = user => {
  return <Link to={`/users/${user.id}`}>{user.name}</Link>
}
