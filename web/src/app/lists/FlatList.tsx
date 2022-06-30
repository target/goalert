import React, { useRef, useState, MouseEvent } from 'react'
import ButtonBase from '@mui/material/ButtonBase'
import IconButton from '@mui/material/IconButton'
import List, { ListProps } from '@mui/material/List'
import MUIListItem from '@mui/material/ListItem'
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction'
import ListItemText from '@mui/material/ListItemText'
import Typography from '@mui/material/Typography'
import ListSubheader from '@mui/material/ListSubheader'
import makeStyles from '@mui/styles/makeStyles'
import { CSSTransition, TransitionGroup } from 'react-transition-group'
import { Alert, AlertTitle } from '@mui/material'
import { AlertColor } from '@mui/material/Alert'
import EditIcon from '@mui/icons-material/Edit'
import {
  closestCenter,
  DndContext,
  DragEndEvent,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
} from '@dnd-kit/sortable'
import classnames from 'classnames'
import { Notice, NoticeType } from '../details/Notices'
import FlatListItem from './FlatListItem'
import { DraggableListItem, getAnnouncements } from './DraggableListItem'

const useStyles = makeStyles({
  alert: {
    margin: '0.5rem 0 0.5rem 0',
    width: '100%',
  },
  alertAsButton: {
    width: '100%',
    '&:hover, &.Mui-focusVisible': {
      filter: 'brightness(90%)',
    },
  },
  buttonBase: {
    borderRadius: 4,
  },
  background: { backgroundColor: 'transparent' },
  listItemText: {
    fontStyle: 'italic',
  },
  slideEnter: {
    maxHeight: '0px',
    opacity: 0,
    transform: 'translateX(-100%)',
  },
  slideEnterActive: {
    maxHeight: '60px',
    opacity: 1,
    transform: 'translateX(0%)',
    transition: 'all 500ms',
  },
  slideExit: {
    maxHeight: '60px',
    opacity: 1,
    transform: 'translateX(0%)',
  },
  slideExitActive: {
    maxHeight: '0px',
    opacity: 0,
    transform: 'translateX(-100%)',
    transition: 'all 500ms',
  },
})

export interface FlatListSub {
  id?: string
  subHeader: JSX.Element | string
}

export interface FlatListNotice extends Notice {
  id?: string
  icon?: JSX.Element
  transition?: boolean
  handleOnClick?: (event: MouseEvent) => void
  'data-cy'?: string
}
export interface FlatListItem {
  title?: string
  highlight?: boolean
  subText?: JSX.Element | string
  icon?: JSX.Element | null
  secondaryAction?: JSX.Element | null
  url?: string
  id?: string // required for drag and drop functionality
  scrollIntoView?: boolean
  'data-cy'?: string
  disabled?: boolean
}

export type FlatListListItem = FlatListSub | FlatListItem | FlatListNotice

export interface FlatListProps extends ListProps {
  items: FlatListListItem[]

  // header elements will be displayed at the top of the list.
  headerNote?: string // left-aligned
  headerAction?: JSX.Element // right-aligned

  // emptyMessage will be displayed if there are no items in the list.
  emptyMessage?: string

  // indent text of each list item if no icon is present
  inset?: boolean

  // If specified, enables drag and drop
  //
  // onReorder(id, oldIndex, newIndex)
  onReorder?: (oldIndex: number, newIndex: number) => void

  // will render transition in list
  transition?: boolean

  // renders an edit button that hides the options buttons until toggled on
  toggleEdit?: boolean
}

const severityMap: { [K in NoticeType]: AlertColor } = {
  INFO: 'info',
  WARNING: 'warning',
  ERROR: 'error',
  OK: 'success',
}

export default function FlatList({
  onReorder,
  emptyMessage,
  headerNote,
  headerAction,
  items,
  inset,
  transition,
  toggleEdit,
  ...listProps
}: FlatListProps): JSX.Element {
  const classes = useStyles()

  // drag and drop stuff
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  )
  const [dndItems, setDndItems] = useState(
    // use IDs to sort, fallback to index
    items.map((i, idx) => (i.id ? i.id : idx.toString())),
  )
  const [canEdit, setCanEdit] = useState(false)
  const isFirstAnnouncement = useRef(false)
  const announcements = getAnnouncements(dndItems, isFirstAnnouncement)
  function handleDragStart(): void {
    if (!isFirstAnnouncement.current) {
      isFirstAnnouncement.current = true
    }
  }
  function handleDragEnd(e: DragEndEvent): void {
    if (!onReorder || !e.over) return
    if (e.active.id !== e.over.id) {
      const oldIndex = dndItems.indexOf(e.active.id.toString())
      const newIndex = dndItems.indexOf(e.over.id.toString())
      setDndItems(arrayMove(dndItems, oldIndex, newIndex)) // update order in local state
      onReorder(oldIndex, newIndex) // callback fn from props
    }
  }

  function renderEmptyMessage(): JSX.Element {
    return (
      <MUIListItem>
        <ListItemText
          disableTypography
          secondary={
            <Typography data-cy='list-empty-message' variant='caption'>
              {emptyMessage}
            </Typography>
          }
        />
      </MUIListItem>
    )
  }

  function renderNoticeItem(item: FlatListNotice, idx: number): JSX.Element {
    if (item.handleOnClick) {
      return (
        <ButtonBase
          className={classnames(classes.buttonBase, classes.alert)}
          onClick={item.handleOnClick}
          data-cy={item['data-cy']}
        >
          <Alert
            className={classes.alertAsButton}
            key={idx}
            severity={severityMap[item.type]}
            icon={item.icon}
          >
            {item.message && <AlertTitle>{item.message}</AlertTitle>}
            {item.details}
          </Alert>
        </ButtonBase>
      )
    }

    return (
      <Alert
        key={idx}
        className={classes.alert}
        severity={severityMap[item.type]}
        icon={item.icon}
      >
        {item.message && <AlertTitle>{item.message}</AlertTitle>}
        {item.details}
      </Alert>
    )
  }

  function renderSubheaderItem(item: FlatListSub, idx: number): JSX.Element {
    return (
      <ListSubheader key={idx} className={classes.background}>
        <Typography
          component='h2'
          variant='subtitle1'
          color='textSecondary'
          data-cy='flat-list-item-subheader'
        >
          {item.subHeader}
        </Typography>
      </ListSubheader>
    )
  }

  function renderTransitionItems(): JSX.Element[] {
    return items.map((item, idx) => {
      if ('subHeader' in item) {
        return (
          <CSSTransition
            key={'header_' + item.id}
            timeout={0}
            exit={false}
            enter={false}
          >
            {renderSubheaderItem(item, idx)}
          </CSSTransition>
        )
      }
      if ('type' in item) {
        return (
          <CSSTransition
            key={'notice_' + item.id}
            timeout={500}
            exit={Boolean(item.transition)}
            enter={Boolean(item.transition)}
            classNames={{
              enter: classes.slideEnter,
              enterActive: classes.slideEnterActive,
              exit: classes.slideExit,
              exitActive: classes.slideExitActive,
            }}
          >
            {renderNoticeItem(item, idx)}
          </CSSTransition>
        )
      }
      return (
        <CSSTransition
          key={'item_' + item.id}
          timeout={500}
          classNames={{
            enter: classes.slideEnter,
            enterActive: classes.slideEnterActive,
            exit: classes.slideExit,
            exitActive: classes.slideExitActive,
          }}
        >
          <FlatListItem
            index={idx}
            item={item}
            canEdit={toggleEdit ? canEdit : true}
          />
        </CSSTransition>
      )
    })
  }

  function renderTransitions(): JSX.Element {
    return <TransitionGroup>{renderTransitionItems()}</TransitionGroup>
  }

  function renderItems(): (JSX.Element | undefined)[] | JSX.Element {
    return items.map((item: FlatListListItem, idx: number) => {
      if ('subHeader' in item) {
        return renderSubheaderItem(item, idx)
      }
      if ('type' in item) {
        return renderNoticeItem(item, idx)
      }
      if (onReorder) {
        return (
          <DraggableListItem
            key={`${idx}-${item.id}`}
            index={idx}
            item={item}
            id={item.id ?? idx.toString()}
            canReorder={toggleEdit ? canEdit : true}
          />
        )
      }

      return (
        <FlatListItem
          key={`${idx}-${item.id}`}
          index={idx}
          item={item}
          canEdit={toggleEdit ? canEdit : true}
        />
      )
    })
  }

  function renderList(): JSX.Element {
    let sx = listProps.sx
    if (onReorder) {
      sx = {
        ...sx,
        display: 'grid',
      }
    }

    return (
      <List {...listProps} sx={sx}>
        {(headerNote || headerAction || onReorder) && (
          <MUIListItem>
            {headerNote && (
              <ListItemText
                disableTypography
                secondary={
                  <Typography color='textSecondary'>{headerNote}</Typography>
                }
                className={classes.listItemText}
              />
            )}
            <div style={{ flex: 1 }} />
            {toggleEdit && (
              <IconButton onClick={() => setCanEdit(!canEdit)}>
                <EditIcon />
              </IconButton>
            )}
            {headerAction && (
              <ListItemSecondaryAction>{headerAction}</ListItemSecondaryAction>
            )}
          </MUIListItem>
        )}
        {!items.length && renderEmptyMessage()}
        {transition ? renderTransitions() : renderItems()}
      </List>
    )
  }

  if (onReorder) {
    return (
      <DndContext
        accessibility={{ announcements }}
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
      >
        <SortableContext items={dndItems}>{renderList()}</SortableContext>
      </DndContext>
    )
  }

  return renderList()
}
