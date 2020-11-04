import React from 'react'
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
import { CSSTransition, TransitionGroup } from 'react-transition-group'

const lime = '#93ed94'
const lightLime = '#defadf'
const lightGrey = '#ebebeb'

const useStyles = makeStyles({
  background: { backgroundColor: 'white' },
  highlightedItem: {
    width: '100%',
    borderLeft: '6px solid ' + lime,
    background: lightLime,
  },
  participantDragging: {
    backgroundColor: lightGrey,
  },
  slideEnter: {
    transform: 'translateX(-100%)',
  },
  slideEnterActive: {
    transform: 'translateX(0%)',
    transition: 'opacity 500ms, transform 500ms',
  },
  slideExit: {
    transform: 'translateX(0%)',
  },
  slideExitActive: {
    transform: 'translateX(-100%)',
    transition: 'opacity 500ms, transform 500ms',
  },
  listItem: {
    width: '100%',
  },
  listItemText: {
    fontStyle: 'italic',
  },
})

export interface FlatListSub {
  subHeader: string
}
export interface FlatListItem {
  title: string
  highlight?: boolean
  subText?: JSX.Element | string
  icon?: JSX.Element
  secondaryAction?: JSX.Element | null
  url?: string
  id?: string
}

export interface ListProps {
  dense: boolean
}

export type FlatListListItem = FlatListSub | FlatListItem

export interface FlatListProps extends Partial<ListProps> {
  items: FlatListListItem[]

  // headerNote will be displayed at the top of the list.
  headerNote?: string

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
}

export default function FlatList({
  onReorder,
  emptyMessage,
  headerNote,
  items,
  inset,
  transition,
  ...listProps
}: FlatListProps): JSX.Element {
  const classes = useStyles()

  function handleDragStart(): void {
    // adds a little vibration if the browser supports it
    if (window.navigator.vibrate) {
      window.navigator.vibrate(100)
    }
  }

  function handleDragEnd(result: DropResult): void {
    if (result.destination && onReorder) {
      onReorder(result.source.index, result.destination.index)
    }
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

  function renderItem(item: FlatListItem, idx: number): JSX.Element {
    let itemProps = {}
    if (item.url) {
      itemProps = {
        component: AppLink,
        to: item.url,
        button: true,
      }
    }
    return (
      <ListItem
        key={idx}
        {...itemProps}
        className={item.highlight ? classes.highlightedItem : classes.listItem}
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

  function renderTransitionItems(): JSX.Element[] {
    return items.map((item, idx) => {
      if ('subHeader' in item) {
        return (
          <CSSTransition key={idx} timeout={0} exit={false} enter={false}>
            {renderSubheaderItem(item, idx)}
          </CSSTransition>
        )
      }
      return (
        <CSSTransition
          key={item.id}
          timeout={500}
          classNames={{
            enter: classes.slideEnter,
            enterActive: classes.slideEnterActive,
            exit: classes.slideExit,
            exitActive: classes.slideExitActive,
          }}
        >
          {renderItem(item, idx)}
        </CSSTransition>
      )
    })
  }

  function renderEmptyMessage(): JSX.Element {
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

  function renderItems(): (JSX.Element | undefined)[] | JSX.Element {
    return items.map((item: FlatListListItem, idx: number) => {
      if ('subHeader' in item) {
        return renderSubheaderItem(item, idx)
      }
      if (!onReorder) {
        return renderItem(item, idx)
      }
      if (item.id) {
        return (
          <Draggable key={item.id} draggableId={item.id} index={idx}>
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
    })
  }

  function renderTransitions(): JSX.Element {
    return <TransitionGroup>{renderTransitionItems()}</TransitionGroup>
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
              className={classes.listItemText}
            />
          </ListItem>
        )}
        {!items.length && renderEmptyMessage()}
        {transition ? renderTransitions() : renderItems()}
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
  return renderList()
}
