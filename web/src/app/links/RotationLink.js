import React from 'react'
import { AppLink } from '../util/AppLink'

export const RotationLink = (rotation) => {
  return <AppLink to={`/rotations/${rotation.id}`}>{rotation.name}</AppLink>
}
