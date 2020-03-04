import React from 'react'
import { AppLink } from '../util/AppLink'

export const ServiceLink = service => {
  return <AppLink to={`/services/${service.id}`}>{service.name}</AppLink>
}
