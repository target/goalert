import React from 'react'
import { FlatListItem as FlatListItemType } from './FlatList'
import FlatListItem from './FlatListItem'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

interface DraggableListItemProps {
  id: string
  index: number
  item: FlatListItemType
  // onReorder: (oldIndex: number, newIndex: number) => void
  // onDrag: (dragIndex: number, hoverIndex: number) => void
}

export function DraggableListItem({
  id,
  index,
  item,
}: DraggableListItemProps): JSX.Element {
  const { attributes, listeners, setNodeRef, transform, transition } =
    useSortable({ id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}>
      <FlatListItem index={index} item={item} />
    </div>
  )
}
