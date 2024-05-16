import React from 'react'
import Chip, { ChipProps } from '@mui/material/Chip'
import { useLocation } from 'wouter'
import { useQuery, gql } from 'urql'

import { ServiceAvatar } from './avatars'

const serviceQuery = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
    }
  }
`

type WithID<T> = { id: string } & T

export function ServiceChip(props: WithID<ChipProps>): JSX.Element {
  const { id, label, ...rest } = props
  const [, navigate] = useLocation()

  const [{ data, fetching, error }] = useQuery({
    query: serviceQuery,
    variables: {
      id,
    },
    pause: Boolean(label),
    requestPolicy: 'cache-first',
  })

  const getLabel = (): typeof label => {
    if (label) return label
    if (!data && fetching) return 'Loading...'
    if (error) return `Error: ${error.message}`
    return data.service.name
  }

  return (
    <Chip
      data-cy='service-chip'
      avatar={<ServiceAvatar />}
      onClick={() => navigate(`/services/${id}`)}
      label={getLabel()}
      {...rest}
    />
  )
}
