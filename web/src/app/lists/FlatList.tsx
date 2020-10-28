import React, { ReactNode } from 'react'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import ListItemText from '@material-ui/core/ListItemText'
import Typography from '@material-ui/core/Typography'
import {
  DragDropContext,
  Droppable,
  Draggable,
  DropResult,
  DraggableStateSnapshot,
  DraggableProvided,
  DroppableProvided,
} from 'react-beautiful-dnd'
import ListSubheader from '@material-ui/core/ListSubheader'
import { AppLink } from '../util/AppLink'
import { makeStyles } from '@material-ui/core'
import {
  CSSTransition,
  TransitionGroup,
  Transition,
} from 'react-transition-group'

const useStyles = makeStyles({
  background: { backgroundColor: 'white' },
  highlightedItem: {
    borderLeft: '6px solid #93ed94',
    background: '#defadf',
  },
  participantDragging: {
    backgroundColor: '#ebebeb',
  },

  fadeEnter: {
    opacity: 0,
    transform: 'translateX(-100%)',
  },
  fadeEnterActive: {
    opacity: 1,
    transform: 'translateX(0%)',
    transition: 'opacity 500ms, transform 500ms',
  },
  fadeExit: {
    opacity: 1,
    transform: 'translateX(0%)',
  },
  fadeExitActive: {
    transform: 'translateX(-100%)',
    transition: 'opacity 500ms, transform 500ms',
  },
})

type FlatListSub = {
  subHeader: string
}
type FlatListItem = {
  highlight?: boolean
  title: string
  subText?: string
  icon?: JSX.Element
  secondaryAction?: JSX.Element | null
  url?: string
  id?: string
}

type FlatListListItem = FlatListSub | FlatListItem

type FlatListType = {
  // headerNote will be displayed at the top of the list.
  headerNote?: ReactNode

  // emptyMessage will be displayed if there are no items in the list.
  emptyMessage?: string

  items: FlatListListItem[]

  listProps?: { dense: boolean }

  // indent text of each list item if no icon is present
  inset?: boolean

  // If specified, enables drag and drop
  //
  // onReorder(id, oldIndex, newIndex)
  onReorder?: (oldIndex: number, newIndex: number) => void

  // will render transition in list
  transition?: boolean
}

export default function FlatList(props: FlatListType): JSX.Element {
  const {
    onReorder,
    emptyMessage,
    headerNote,
    items,
    inset,
    transition,
    listProps,
  } = props

  const classes = useStyles()

  const handleDragStart = (): void => {
    // adds a little vibration if the browser supports it
    if (window.navigator.vibrate) {
      window.navigator.vibrate(100)
    }
  }

  const handleDragEnd = (result: DropResult): void => {
    if (result.destination) {
      if (onReorder) {
        onReorder(
          // result.draggableId, : removed this as per new reorderList function
          result.source.index,
          result.destination.index,
        )
      }
    }
  }

  function renderItem(item: FlatListItem, idx: number): JSX.Element {
    let itemProps = {}
    if (item.url) {
      itemProps = {
        component: AppLink,
        to: item.url,
        button: true,
      }
    }
    if (transition) {
      return (
        <CSSTransition
          key={item.id}
          timeout={500}
          classNames={{
            enter: classes.fadeEnter,
            enterActive: classes.fadeEnterActive,
            exit: classes.fadeExit,
            exitActive: classes.fadeExitActive,
          }}
        >
          <ListItem
            key={idx}
            {...itemProps}
            style={{ width: '100%' }}
            className={item.highlight ? classes.highlightedItem : ''}
          >
            {item.icon && <ListItemIcon>{item.icon}</ListItemIcon>}
            <ListItemText
              primary={item.title}
              secondary={item.subText}
              secondaryTypographyProps={{ style: { whiteSpace: 'pre-line' } }}
              inset={inset && !item.icon}
            />
            {item.secondaryAction && (
              <ListItemSecondaryAction>
                {item.secondaryAction}
              </ListItemSecondaryAction>
            )}
          </ListItem>
        </CSSTransition>
      )
    }

    return (
      <ListItem
        key={idx}
        {...itemProps}
        style={{ width: '100%' }}
        className={item.highlight ? classes.highlightedItem : ''}
      >
        {item.icon && <ListItemIcon>{item.icon}</ListItemIcon>}
        <ListItemText
          primary={item.title}
          secondary={item.subText}
          secondaryTypographyProps={{ style: { whiteSpace: 'pre-line' } }}
          inset={inset && !item.icon}
        />
        {item.secondaryAction && (
          <ListItemSecondaryAction>
            {item.secondaryAction}
          </ListItemSecondaryAction>
        )}
      </ListItem>
    )
  }

  function renderItems(): (JSX.Element | undefined)[] | JSX.Element {
    if (!items.length) {
      return (
        <ListItem>
          <ListItemText
            disableTypography
            secondary={
              <Typography data-cy='list-empty-message' variant='caption'>
                {emptyMessage}
              </Typography>
            }
          />
        </ListItem>
      )
    }

    return items.map((item: FlatListListItem, idx: number) => {
      if (!onReorder) {
        if ('subHeader' in item) {
          if (item.subHeader) {
            if (transition) {
              return (
                <Transition timeout={500}>
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
                </Transition>
              )
            }
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
        } else {
          return renderItem(item, idx)
        }
      }
      if ('id' in item) {
        if (item.id) {
          return (
            <Draggable key={idx + item.id} draggableId={item.id} index={idx}>
              {(
                provided: DraggableProvided,
                snapshot: DraggableStateSnapshot,
              ) => {
                // light grey background while dragging non-active user
                const draggingBackground = snapshot.isDragging
                  ? classes.participantDragging
                  : ''
                return (
                  <div
                    ref={provided.innerRef}
                    {...provided.draggableProps}
                    {...provided.dragHandleProps}
                    className={draggingBackground}
                  >
                    {renderItem(item, idx)}
                  </div>
                )
              }}
            </Draggable>
          )
        }
      }
    })
  }

  function renderList(): JSX.Element {
    return (
      <List {...listProps}>
        {headerNote && (
          <ListItem>
            <ListItemText
              disableTypography
              secondary={
                <Typography color='textSecondary'>{headerNote}</Typography>
              }
              style={{ fontStyle: 'italic' }}
            />
          </ListItem>
        )}
        {renderItems()}
      </List>
    )
  }

  function renderTransitionList(): JSX.Element {
    return (
      <List {...listProps}>
        <TransitionGroup>
          {headerNote && (
            <ListItem>
              <ListItemText
                disableTypography
                secondary={
                  <Typography color='textSecondary'>{headerNote}</Typography>
                }
                style={{ fontStyle: 'italic' }}
              />
            </ListItem>
          )}
          {renderItems()}
        </TransitionGroup>
      </List>
    )
  }

  function renderDragAndDrop(): JSX.Element {
    return (
      <DragDropContext onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
        <Droppable droppableId='droppable'>
          {(provided: DroppableProvided) => (
            <div ref={provided.innerRef} {...provided.droppableProps}>
              {renderList()}
              {provided.placeholder}
            </div>
          )}
        </Droppable>
      </DragDropContext>
    )
  }

  if (onReorder) {
    // Enable drag and drop
    return renderDragAndDrop()
  }
  if (transition) {
    return renderTransitionList()
  }
  return renderList()
}
