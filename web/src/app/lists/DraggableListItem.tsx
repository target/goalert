import React from 'react'
import { FlatListItem as FlatListItemType } from './FlatList'
import FlatListItem from './FlatListItem'
import { CSS } from '@dnd-kit/utilities'

import {
  // arrayMove,
  useSortable,
  // SortableContext,
  // sortableKeyboardCoordinates,
  // SortingStrategy,
  // rectSortingStrategy,
  // AnimateLayoutChanges,
  // NewIndexGetter,
} from '@dnd-kit/sortable'
import { useTheme } from '@mui/material'

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
    active,
    attributes,
    isDragging,
    isSorting,
    listeners,
    overIndex,
    setNodeRef,
    setActivatorNodeRef,
    transform,
    transition,
  } = useSortable({
    id,
    // animateLayoutChanges,
    // disabled,
    // getNewIndex,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    backgroundColor: isDragging ? theme.palette.background.default : 'inherit',
  }

  // todo: style when dragging
  // todo: show different mouse pointer when hovering
  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}>
      <FlatListItem index={index} item={item} />
    </div>
  )
}
