import React from 'react'
import { Link } from 'react-router-dom'

export const ServiceLink = service => {
  return <Link to={`/services/${service.id}`}>{service.name}</Link>
}
