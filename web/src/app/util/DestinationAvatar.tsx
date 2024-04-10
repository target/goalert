import React from 'react'
import { Avatar, CircularProgress } from '@mui/material'

import {
  BrokenImage,
  Notifications,
  RotateRight as RotationIcon,
  Today as ScheduleIcon,
  Webhook as WebhookIcon,
} from '@mui/icons-material'

const builtInIcons: { [key: string]: React.ReactNode } = {
  'builtin://rotation': <RotationIcon />,
  'builtin://schedule': <ScheduleIcon />,
  'builtin://webhook': <WebhookIcon />,
  'builtin://alert': <Notifications />,
}

export type DestinationAvatarProps = {
  error?: boolean
  loading?: boolean
  iconURL?: string
  iconAltText?: string
}

/**
 * DestinationAvatar is used to display the icon for a selected destination value.
 *
 * It will return null if the iconURL is not provided, and there is no error or loading state.
 */
export function DestinationAvatar(
  props: DestinationAvatarProps,
): React.ReactNode {
  const { error, loading, iconURL, iconAltText, ...rest } = props

  if (props.error) {
    return (
      <Avatar {...rest}>
        <BrokenImage />
      </Avatar>
    )
  }

  if (loading) {
    return (
      <Avatar {...rest}>
        <CircularProgress data-testid='spinner' size='1em' />
      </Avatar>
    )
  }

  if (!iconURL) {
    return null
  }

  const builtInIcon = builtInIcons[iconURL] || null
  return (
    <Avatar {...rest} src={builtInIcon ? undefined : iconURL} alt={iconAltText}>
      {builtInIcon}
    </Avatar>
  )
}
