import React, { MutableRefObject } from 'react'
import { FlatListItem as FlatListItemType } from './FlatList'
import FlatListItem from './FlatListItem'
import { Announcements, UniqueIdentifier } from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import { useSortable } from '@dnd-kit/sortable'
import { useTheme } from '@mui/material'

export function getAnnouncements(
  items: string[],
  isFirstAnnouncement: MutableRefObject<boolean>,
): Announcements {
  const getPosition = (id: UniqueIdentifier): number =>
    items.indexOf(id.toString()) + 1

  return {
    onDragStart({ active: { id } }) {
      return `Picked up sortable item ${String(
        id,
      )}. Sortable item ${id} is in position ${getPosition(id)} of ${
        items.length
      }`
    },
    onDragOver({ active, over }) {
      // onDragOver is called right after onDragStart, cancel first run here
      // in favor of the pickup announcement
      if (isFirstAnnouncement.current) {
        isFirstAnnouncement.current = false
        return
      }

      if (over) {
        return `Sortable item ${
          active.id
        } was moved into position ${getPosition(over.id)} of ${items.length}`
      }
    },
    onDragEnd({ active, over }) {
      if (over) {
        return `Sortable item ${
          active.id
        } was dropped at position ${getPosition(over.id)} of ${items.length}`
      }
    },
    onDragCancel({ active: { id } }) {
      return `Sorting was cancelled. Sortable item ${id} was dropped and returned to position ${getPosition(
        id,
      )} of ${items.length}.`
    },
  }
}

interface DraggableListItemProps {
  id: string
  index: number
  item: FlatListItemType
}

export function DraggableListItem({
  id,
  index,
  item,
}: DraggableListItemProps): JSX.Element {
  const theme = useTheme()
  const {
    attributes,
    isDragging,
    listeners,
    setNodeRef,
    transform,
    transition,
  } = useSortable({
    id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    backgroundColor: isDragging ? theme.palette.background.default : 'inherit',
    cursor: 'pointer',
    zIndex: isDragging ? 9001 : 1,
  }

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}>
      <FlatListItem index={index} item={item} />
    </div>
  )
}
