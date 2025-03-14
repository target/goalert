import React from 'react'
import {
  Announcements,
  closestCenter,
  DndContext,
  DragEndEvent,
  KeyboardSensor,
  MeasuringStrategy,
  PointerSensor,
  UniqueIdentifier,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import { ReorderableItemProps } from './ReorderableItem'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'

const measuringConfig = {
  droppable: {
    strategy: MeasuringStrategy.Always,
  },
}

export function getAnnouncements(
  items: string[],
  isFirstAnnouncement: React.MutableRefObject<boolean>,
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

export type DraggableCompListProps = {
  onReorder: (from: number, to: number) => void
  children: React.ReactElement<ReorderableItemProps>[]
}

/* A list of draggable components. */
export default function ReorderGroup(
  props: DraggableCompListProps,
): React.ReactNode {
  // collect keys from children
  const [dndItems, setDndItems] = React.useState(
    React.Children.map(props.children, (child) => {
      return child.props.id
    }),
  )
  React.useEffect(() => {
    setDndItems(
      React.Children.map(props.children, (child) => {
        return child.props.id
      }),
    )
  }, [props.children])
  const isFirstAnnouncement = React.useRef(false)
  const announcements = getAnnouncements(dndItems, isFirstAnnouncement)
  function handleDragStart(): void {
    if (!isFirstAnnouncement.current) {
      isFirstAnnouncement.current = true
    }
  }
  function handleDragEnd(e: DragEndEvent): void {
    if (!e.over) return
    if (e.active.id === e.over.id) return
    const oldIndex = dndItems.indexOf(e.active.id.toString())
    const newIndex = dndItems.indexOf(e.over.id.toString())
    setDndItems(arrayMove(dndItems, oldIndex, newIndex)) // update order in local state
    props.onReorder(oldIndex, newIndex) // callback fn from props
  }
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  )

  return (
    <DndContext
      accessibility={{ announcements }}
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      measuring={measuringConfig}
    >
      <SortableContext items={dndItems} strategy={verticalListSortingStrategy}>
        {props.children}
      </SortableContext>
    </DndContext>
  )
}
