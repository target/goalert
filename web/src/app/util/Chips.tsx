import React from 'react'
import Chip, { ChipProps } from '@mui/material/Chip'
import { useLocation } from 'wouter'
import { useQuery, gql } from 'urql'
import {
  RotateRight as RotationIcon,
  Today as ScheduleIcon,
  Webhook as WebhookIcon,
} from '@mui/icons-material'
import Avatar from '@mui/material/Avatar'

import { UserAvatar, ServiceAvatar } from './avatars'
import { SlackBW } from '../icons'
import { Query } from '../../schema'

const serviceQuery = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
    }
  }
`

type WithID<T> = { id: string } & T

export function ServiceChip(props: WithID<ChipProps>): React.ReactNode {
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

export function UserChip(props: WithID<ChipProps>): React.ReactNode {
  const { id, ...rest } = props
  const [, navigate] = useLocation()

  return (
    <Chip
      data-cy='user-chip'
      avatar={<UserAvatar userID={id} />}
      onClick={() => navigate(`/users/${id}`)}
      {...rest}
    />
  )
}

export function RotationChip(props: WithID<ChipProps>): React.ReactNode {
  const { id, ...rest } = props
  const [, navigate] = useLocation()

  return (
    <Chip
      data-cy='rotation-chip'
      avatar={
        <Avatar>
          <RotationIcon />
        </Avatar>
      }
      onClick={() => navigate(`/rotations/${id}`)}
      {...rest}
    />
  )
}

export function ScheduleChip(props: WithID<ChipProps>): React.ReactNode {
  const { id, ...rest } = props
  const [, navigate] = useLocation()

  return (
    <Chip
      data-cy='schedule-chip'
      avatar={
        <Avatar>
          <ScheduleIcon />
        </Avatar>
      }
      onClick={() => navigate(`/schedules/${id}`)}
      {...rest}
    />
  )
}

export function SlackChip(props: WithID<ChipProps>): React.ReactNode {
  const { id: channelID, ...rest } = props

  const query = gql`
    query ($id: ID!) {
      slackChannel(id: $id) {
        id
        teamID
      }
    }
  `

  const [{ data, error }] = useQuery<Query>({
    query,
    variables: { id: channelID },
    requestPolicy: 'cache-first',
  })
  const teamID = data?.slackChannel?.teamID

  if (error) {
    console.error(`Error querying slackChannel ${channelID}:`, error)
  }

  const clickable = Boolean(channelID && teamID)
  if (clickable) {
    rest.onClick = () =>
      window.open(
        `https://slack.com/app_redirect?channel=${channelID}&team=${teamID}`,
      )
  }

  return (
    <Chip
      data-cy='slack-chip'
      data-clickable={clickable}
      avatar={
        <Avatar>
          <SlackBW />
        </Avatar>
      }
      {...rest}
    />
  )
}

export function WebhookChip(props: WithID<ChipProps>): React.ReactNode {
  const { id: urlStr, ...rest } = props

  const url = new URL(urlStr)
  return (
    <Chip
      data-cy='webhook-chip'
      avatar={
        <Avatar>
          <WebhookIcon />
        </Avatar>
      }
      label={url.hostname}
      {...rest}
    />
  )
}
