import React from 'react'
import Chip, { ChipProps } from '@material-ui/core/Chip'
import { push } from 'connected-react-router'
import { useDispatch } from 'react-redux'
import { useQuery, gql } from '@apollo/client'
import {
  RotateRight as RotationIcon,
  Today as ScheduleIcon,
} from '@material-ui/icons'
import Avatar from '@material-ui/core/Avatar'

import { UserAvatar, ServiceAvatar } from './avatars'
import { SlackBW } from '../icons'

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
  const { id, label, onClick, ...rest } = props
  const dispatch = useDispatch()

  const { data, loading, error } = useQuery(serviceQuery, {
    variables: {
      id,
    },
    skip: Boolean(label),
    fetchPolicy: 'cache-first',
    pollInterval: 0,
  })

  const getLabel = (): typeof label => {
    if (label) return label
    if (!data && loading) return 'Loading...'
    if (error) return `Error: ${error.message}`
    return data.service.name
  }

  return (
    <Chip
      data-cy='service-chip'
      avatar={<ServiceAvatar />}
      onClick={onClick || (() => dispatch(push(`/services/${id}`)))}
      label={getLabel()}
      {...rest}
    />
  )
}

export function UserChip(props: WithID<ChipProps>): JSX.Element {
  const { id, onClick, ...rest } = props
  const dispatch = useDispatch()

  return (
    <Chip
      data-cy='user-chip'
      avatar={<UserAvatar userID={id} />}
      onClick={onClick || (() => dispatch(push(`/users/${id}`)))}
      {...rest}
    />
  )
}

export function RotationChip(props: WithID<ChipProps>): JSX.Element {
  const { id, onClick, ...rest } = props
  const dispatch = useDispatch()

  return (
    <Chip
      data-cy='rotation-chip'
      avatar={
        <Avatar>
          <RotationIcon />
        </Avatar>
      }
      onClick={onClick || (() => dispatch(push(`/rotations/${id}`)))}
      {...rest}
    />
  )
}

export function ScheduleChip(props: WithID<ChipProps>): JSX.Element {
  const { id, onClick, ...rest } = props
  const dispatch = useDispatch()

  return (
    <Chip
      data-cy='schedule-chip'
      avatar={
        <Avatar>
          <ScheduleIcon />
        </Avatar>
      }
      onClick={onClick || (() => dispatch(push(`/schedules/${id}`)))}
      {...rest}
    />
  )
}

export function SlackChip(props: ChipProps): JSX.Element {
  return (
    <Chip
      data-cy='slack-chip'
      avatar={
        <Avatar>
          <SlackBW />
        </Avatar>
      }
      {...props}
    />
  )
}
