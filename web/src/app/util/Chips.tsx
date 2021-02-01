import React from 'react'
import Chip, { ChipProps } from '@material-ui/core/Chip'
import { push } from 'connected-react-router'
import { useDispatch } from 'react-redux'
import { useQuery, gql } from '@apollo/client'
import {
  Layers as PolicyIcon,
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

interface ServiceChipProps extends ChipProps {
  id: string
  name?: string
}

export function ServiceChip(props: ServiceChipProps): JSX.Element {
  const { id, name, onClick, ...rest } = props
  const dispatch = useDispatch()

  const { data, loading, error } = useQuery(serviceQuery, {
    variables: {
      id,
    },
    skip: Boolean(name),
    fetchPolicy: 'cache-first',
    pollInterval: 0,
  })

  const getLabel = (): string => {
    if (name) return name
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

interface UserChipProps extends ChipProps {
  id: string
  name: string
}

export function UserChip(props: UserChipProps): JSX.Element {
  const { id, name, onClick, ...rest } = props
  const dispatch = useDispatch()

  return (
    <Chip
      data-cy='user-chip'
      avatar={<UserAvatar userID={id} />}
      onClick={onClick || (() => dispatch(push(`/users/${id}`)))}
      label={name}
      {...rest}
    />
  )
}

interface RotationChipProps extends ChipProps {
  id: string
  name: string
}

export function RotationChip(props: RotationChipProps): JSX.Element {
  const { id, name, onClick, ...rest } = props
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
      label={name}
      {...rest}
    />
  )
}

function formatPolicyName(name: string, stepNum: number): JSX.Element {
  if (stepNum !== null) {
    return (
      <div>
        <strong> Step {stepNum + 1}:</strong> {name}
      </div>
    )
  }

  const parts = name.split(' - ')
  return (
    <div>
      <strong> Step {parseInt(parts[0]) + 1}:</strong> {parts[1]}
    </div>
  )
}

interface PolicyChipProps extends ChipProps {
  id: string
  name: string
  stepNum: number
}

export function PolicyChip(props: PolicyChipProps): JSX.Element {
  const { id, name, stepNum, onClick, ...rest } = props
  const dispatch = useDispatch()

  return (
    <Chip
      data-cy='ep-chip'
      avatar={
        <Avatar>
          <PolicyIcon />
        </Avatar>
      }
      onClick={onClick || (() => dispatch(push(`/escalation-policies/${id}`)))}
      label={formatPolicyName(name, stepNum)}
      {...rest}
    />
  )
}

interface ScheduleChipProps extends ChipProps {
  id: string
  name: string
}

export function ScheduleChip(props: ScheduleChipProps): JSX.Element {
  const { id, name, onClick, ...rest } = props
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
      label={name}
      {...rest}
    />
  )
}

interface SlackChipProps extends ChipProps {
  name: string
}

export function SlackChip(props: SlackChipProps): JSX.Element {
  const { name, ...rest } = props

  return (
    <Chip
      data-cy='slack-chip'
      avatar={
        <Avatar>
          <SlackBW />
        </Avatar>
      }
      label={name}
      {...rest}
    />
  )
}
