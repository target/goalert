import React from 'react'
import { Chip } from '@mui/material'
import { InlineDisplayInfo } from '../../schema'
import { DestinationAvatar } from './DestinationAvatar'

export type DestinationChipProps = InlineDisplayInfo & {
  // If onDelete is provided, a delete icon will be shown.
  onDelete?: () => void
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
  if ('error' in props) {
    return (
      <Chip
        avatar={<DestinationAvatar error />}
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
        avatar={<DestinationAvatar loading />}
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
        <DestinationAvatar
          iconURL={props.iconURL}
          iconAltText={props.iconAltText}
        />
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
