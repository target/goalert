import React from 'react'
import { Service } from '../../schema'
import AppLink from '../util/AppLink'

export const ServiceLink = (service: Service): JSX.Element => {
  return <AppLink to={`/services/${service.id}`}>{service.name}</AppLink>
}
