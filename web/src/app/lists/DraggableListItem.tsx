import React, { useRef } from 'react'
import { useDrag, useDrop } from 'react-dnd'
import { FlatListItem } from './FlatList'
import ListItem from './ListItem'

export const ItemTypes = {
  ITEM: 'item',
}

interface DraggableListItemProps extends Omit<FlatListItem, 'id'> {
  id: number
  index: number
  item: FlatListItem
  onReorder: (oldIndex: number, newIndex: number) => void
  onDrag: (dragIndex: number, hoverIndex: number) => void
}

interface DragItem {
  index: number
  id: string
  type: string
}

export function DraggableListItem({
  id,
  index,
  item,
  onDrag,
  onReorder,
}: DraggableListItemProps): JSX.Element {
  const ref = useRef<HTMLDivElement>(null)
  const [{ handlerId }, drop] = useDrop<
    DragItem,
    void,
    { handlerId: string | symbol | null }
  >({
    accept: ItemTypes.ITEM,
    collect(monitor) {
      return {
        handlerId: monitor.getHandlerId(),
      }
    },
    // update local state item here
    hover(item: DragItem, monitor) {
      if (!ref.current) {
        return
      }
      const dragIndex = item.index
      const hoverIndex = index
      // Don't replace items with themselves
      if (dragIndex === hoverIndex) {
        return
      }
      // Determine rectangle on screen
      const hoverBoundingRect = ref.current?.getBoundingClientRect()
      // Get vertical middle
      const hoverMiddleY =
        (hoverBoundingRect.bottom - hoverBoundingRect.top) / 2
      // Determine mouse position
      const clientOffset = monitor.getClientOffset()
      // Get pixels to the top
      const hoverClientY = clientOffset?.y ?? 0 - hoverBoundingRect.top
      // Only perform the move when the mouse has crossed half of the items height
      // When dragging downwards, only move when the cursor is below 50%
      // When dragging upwards, only move when the cursor is above 50%
      // Dragging downwards
      if (dragIndex < hoverIndex && hoverClientY < hoverMiddleY) {
        return
      }
      // Dragging upwards
      if (dragIndex > hoverIndex && hoverClientY > hoverMiddleY) {
        return
      }
      // Time to actually perform the action
      onDrag(dragIndex, hoverIndex)
      // Note: we're mutating the monitor item here!
      // Generally it's better to avoid mutations,
      // but it's good here for the sake of performance
      // to avoid expensive index searches.
      item.index = hoverIndex
    },
  })

  const [{ isDragging }, drag] = useDrag({
    type: ItemTypes.ITEM,
    item: () => {
      return { id, index }
    },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
    // update state in backend on final "drop" here
    end: ({ id: oldIndex, index: newIndex }) => {
      onReorder(oldIndex, newIndex)
    },
  })

  drag(drop(ref))
  return (
    <div
      ref={ref}
      style={{ cursor: 'move', opacity: isDragging ? 0 : 1 }}
      data-handler-id={handlerId}
    >
      <ListItem index={index} item={item} />
    </div>
  )
}
