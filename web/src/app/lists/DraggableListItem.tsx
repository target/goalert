import React, { MutableRefObject } from 'react'
import { FlatListItem as FlatListItemType } from './FlatList'
import FlatListItem from './FlatListItem'
import { Announcements, UniqueIdentifier } from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import {
  useSortable,
  defaultAnimateLayoutChanges,
  AnimateLayoutChanges,
} from '@dnd-kit/sortable'
import { Button, useTheme } from '@mui/material'
import DragHandleIcon from '@mui/icons-material/DragHandle'

export function getAnnouncements(
  items: string[],
  isFirstAnnouncement: MutableRefObject<boolean>,
): Announcements {
  const getPosition = (id: UniqueIdentifier): number => {
    return items.indexOf(id.toString()) + 1
  }

  return {
    onDragStart({ active: { id } }) {
      return `Picked up sortable item ${getPosition(
        id,
      )}. Sortable item ${getPosition(id)} is in position ${getPosition(
        id,
      )} of ${items.length}`
    },
    onDragOver({ active, over }) {
      // onDragOver is called right after onDragStart, cancel first run here
      // in favor of the pickup announcement
      if (isFirstAnnouncement.current) {
        isFirstAnnouncement.current = false
        return
      }

      if (over) {
        return `Sortable item ${getPosition(
          active.id,
        )} was moved into position ${getPosition(over.id)} of ${items.length}`
      }
    },
    onDragEnd({ active, over }) {
      if (over) {
        return `Sortable item ${getPosition(
          active.id,
        )} was dropped at position ${getPosition(over.id)} of ${items.length}`
      }
    },
    onDragCancel({ active: { id } }) {
      return `Sorting was cancelled. Sortable item ${id} was dropped and returned to position ${getPosition(
        id,
      )} of ${items.length}.`
    },
  }
}

const animateLayoutChanges: AnimateLayoutChanges = (args) =>
  args.isSorting || args.wasDragging ? defaultAnimateLayoutChanges(args) : true

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
    animateLayoutChanges,
    id,
  })

  return (
    <div
      ref={setNodeRef}
      style={{
        display: 'flex',
        transform: CSS.Translate.toString(transform),
        transition,
        backgroundColor: isDragging
          ? theme.palette.background.default
          : 'inherit',
        zIndex: isDragging ? 9001 : 1,
      }}
      {...attributes}
    >
      <div
        style={{
          position: 'absolute',
          padding: '20px',
          zIndex: 2,
        }}
      >
        <Button
          variant='outlined'
          id={'drag-' + index}
          sx={{
            p: '2px',
            width: 'fit-content',
            height: 'fit-content',
            minWidth: 0,
            cursor: 'drag',
          }}
          {...listeners}
        >
          <DragHandleIcon />
        </Button>
      </div>

      <div style={{ width: '100%' }}>
        <FlatListItem index={index} item={{ ...item, draggable: true }} />
      </div>
    </div>
  )
}
