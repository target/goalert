import React from 'react'
import { Chip } from '@mui/material'
import { InlineDisplayInfo } from '../../schema'
import { DestinationAvatar } from './DestinationAvatar'
import { Edit } from '@mui/icons-material'

export type DestinationChipProps = InlineDisplayInfo & {
  // If onDelete is provided, a delete icon will be shown.
  onDelete?: () => void

  /* if onEdit is provided, an edit icon will be shown. Takes precedence over onDelete */
  onEdit?: () => void
  onChipClick?: () => void
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
  const deleteIcon = props.onEdit ? <Edit /> : undefined
  const onDel = props.onEdit || props.onDelete
  const handleOnDelete = onDel
    ? (e: React.MouseEvent) => {
        e.stopPropagation()
        e.preventDefault()
        onDel()
      }
    : undefined

  if ('error' in props) {
    return (
      <Chip
        avatar={<DestinationAvatar error />}
        label={'ERROR: ' + props.error}
        onDelete={handleOnDelete}
        deleteIcon={deleteIcon}
      />
    )
  }
  if (!props.text) {
    return (
      <Chip
        avatar={<DestinationAvatar loading />}
        label='Loading...'
        onDelete={handleOnDelete}
        deleteIcon={deleteIcon}
      />
    )
  }

  const opts: { [key: string]: unknown } = {}
  if (props.linkURL && !props.onChipClick) {
    opts.href = props.linkURL
    opts.target = '_blank'
    opts.component = 'a'
    opts.rel = 'noopener noreferrer'
  }

  return (
    <Chip
      data-testid='destination-chip'
      clickable={!!props.linkURL || !!props.onChipClick}
      onClick={props.onChipClick}
      {...opts}
      avatar={
        <DestinationAvatar
          iconURL={props.iconURL}
          iconAltText={props.iconAltText}
        />
      }
      label={props.text}
      onDelete={handleOnDelete}
      deleteIcon={deleteIcon}
    />
  )
}
