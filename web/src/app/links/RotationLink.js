import React from 'react'
import { MuiLink } from '../util/AppLink'

export const RotationLink = (rotation) => {
  return <MuiLink to={`/rotations/${rotation.id}`}>{rotation.name}</MuiLink>
}
