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
    hover(item: DragItem) {
      if (!ref.current) return
      if (item.index === index) return
      onDrag(item.index, index)
      item.index = index
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
