import React from 'react'
import { MuiLink } from '../util/AppLink'

export const ServiceLink = (service) => {
  return <MuiLink to={`/services/${service.id}`}>{service.name}</MuiLink>
}
