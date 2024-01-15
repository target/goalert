import React from 'react'
import { Avatar, Chip, CircularProgress } from '@mui/material'

import {
  BrokenImage,
  RotateRight as RotationIcon,
  Today as ScheduleIcon,
  Webhook as WebhookIcon,
} from '@mui/icons-material'
import { DestinationDisplayInfo } from '../../schema'

export type DestinationChipProps = DestinationDisplayInfo & {
  error?: string

  // If onDelete is provided, a delete icon will be shown.
  onDelete?: () => void
}

const builtInIcons: { [key: string]: React.ReactNode } = {
  'builtin://rotation': <RotationIcon />,
  'builtin://schedule': <ScheduleIcon />,
  'builtin://webhook': <WebhookIcon />,
}

/**
 * DestinationChip is used to display a selected destination value.
 *
 * You should almost never use this component directly. Instead, use
 * DestinationInputChip, which will select the correct values based on the
 * provided DestinationInput value.
 */
export default function DestinationChip(
  props: DestinationChipProps,
): React.ReactNode {
  if (props.error) {
    return (
      <Chip
        avatar={
          <Avatar>
            <BrokenImage />
          </Avatar>
        }
        label={'ERROR: ' + props.error}
        onDelete={
          props.onDelete
            ? (e) => {
                e.stopPropagation()
                e.preventDefault()
                props.onDelete?.()
              }
            : undefined
        }
      />
    )
  }
  if (!props.text) {
    return (
      <Chip
        avatar={
          <Avatar>
            <CircularProgress data-testid='spinner' size='1em' />
          </Avatar>
        }
        label='Loading...'
        onDelete={
          props.onDelete
            ? (e) => {
                e.stopPropagation()
                e.preventDefault()
                props.onDelete?.()
              }
            : undefined
        }
      />
    )
  }

  const builtInIcon = builtInIcons[props.iconURL] || null

  const opts: { [key: string]: unknown } = {}
  if (props.linkURL) {
    opts.href = props.linkURL
    opts.target = '_blank'
    opts.component = 'a'
    opts.rel = 'noopener noreferrer'
  }

  return (
    <Chip
      data-testid='destination-chip'
      clickable={!!props.linkURL}
      {...opts}
      avatar={
        props.iconURL ? (
          <Avatar
            src={builtInIcon ? undefined : props.iconURL}
            alt={props.iconAltText}
          >
            {builtInIcon}
          </Avatar>
        ) : undefined
      }
      label={props.text}
      onDelete={
        props.onDelete
          ? (e) => {
              e.stopPropagation()
              e.preventDefault()
              props.onDelete?.()
            }
          : undefined
      }
    />
  )
}
