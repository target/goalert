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
import { Grid, ButtonGroup, Button, useTheme } from '@mui/material'
import DragHandleIcon from '@mui/icons-material/DragHandle'
import { ChevronDown, ChevronUp } from 'mdi-material-ui'

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

  const style = {
    transform: CSS.Translate.toString(transform),
    transition,
    backgroundColor: isDragging ? theme.palette.background.default : 'inherit',
  }

  return (
    <Grid container ref={setNodeRef} style={style} {...attributes}>
      <Grid item sx={{ display: 'flex', alignItems: 'center' }}>
        <Button
          variant='outlined'
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
      </Grid>

      <Grid
        item
        xs={11}
        sx={{
          width: '100%',
        }}
      >
        <FlatListItem index={index} item={item} />
      </Grid>

      <Grid item sx={{ display: 'flex', alignItems: 'center' }}>
        <ButtonGroup orientation='vertical'>
          <Button sx={{ p: '2px', width: 'fit-content', minWidth: 0 }}>
            <ChevronUp />
          </Button>
          <Button sx={{ p: '2px', width: 'fit-content', minWidth: 0 }}>
            <ChevronDown />
          </Button>
        </ButtonGroup>
      </Grid>
    </Grid>
  )
}
