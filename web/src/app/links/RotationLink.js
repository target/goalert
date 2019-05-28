import React from 'react'
import { Link } from 'react-router-dom'

export const RotationLink = rotation => {
  return <Link to={`/rotations/${rotation.id}`}>{rotation.name}</Link>
}
