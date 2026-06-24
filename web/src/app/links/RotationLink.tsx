import React from 'react'
import AppLink from '../util/AppLink'
import { Target } from '../../schema'

export const RotationLink = (rotation: Target): React.JSX.Element => {
  return <AppLink to={`/rotations/${rotation.id}`}>{rotation.name}</AppLink>
}
