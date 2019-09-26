import React from 'react'
import p from 'prop-types'
import classnames from 'classnames'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import ListItemText from '@material-ui/core/ListItemText'
import Typography from '@material-ui/core/Typography'
import { DragDropContext, Droppable, Draggable } from 'react-beautiful-dnd'
import ListSubheader from '@material-ui/core/ListSubheader'
import { Link } from 'react-router-dom'
import { absURLSelector } from '../selectors'
import { useSelector } from 'react-redux'
import { makeStyles } from '@material-ui/core'

const useStyles = makeStyles({
  background: {
    backgroundColor: 'white',
  },
  dndDragging: {
    backgroundColor: '#ebebeb',
  },
  highlightedItem: {
    borderLeft: '6px solid #93ed94',
    background: '#defadf',
  },
  listSubheader: {
    margin: 0,
  },
  listItem: {
    width: '100%',
  },
  listItemSubtext: {
    whiteSpace: 'pre-line',
  },
})

function onDragStart() {
  // adds a little vibration if the browser supports it
  if (window.navigator.vibrate) {
    window.navigator.vibrate(100)
  }
}

/**
 * This component will render a simple list on the page.
 * Drag and dropping of items is supported if an "onReorder"
 * function is provided. Other options are listed in the
 * propTypes.
 */
export default function FlatList(props) {
  const classes = useStyles()
  const absURL = useSelector(absURLSelector)

  return props.onReorder ? renderDragAndDropList() : renderList()

  function renderDragAndDropList() {
    return (
      <DragDropContext onDragStart={onDragStart} onDragEnd={props.onReorder}>
        <Droppable droppableId='droppable'>
          {(provided, _) => (
            <div ref={provided.innerRef} {...provided.droppableProps}>
              {renderList(provided.placeholder)}
            </div>
          )}
        </Droppable>
      </DragDropContext>
    )
  }

  function renderList(dragPlaceholder) {
    const { onReorder, emptyMessage, headerNote, items, ...otherProps } = props

    const subheader = headerNote ? (
      <ListSubheader component='p' className={classes.listSubheader}>
        {headerNote}
      </ListSubheader>
    ) : null

    return (
      <List subheader={subheader} {...otherProps}>
        {renderListItems()}

        {/* Rendered if props.onReorder is specified */}
        {dragPlaceholder}
      </List>
    )
  }

  /*
   * Handles rendering either as empty message text,
   * standard list items, list items that support
   * drag and drop, or as a subheader list item.
   */
  function renderListItems() {
    // render as empty message
    if (!props.items.length) {
      return <ListSubheader>{props.emptyMessage}</ListSubheader>
    }

    return props.items.map((item, idx) => {
      // render with drag and drop
      if (props.onReorder) {
        return (
          <Draggable key={idx + item.id} draggableId={item.id} index={idx}>
            {(provided, snapshot) => {
              // light grey background while dragging non-active user
              const draggingBackground = snapshot.isDragging
                ? classes.dndDragging
                : null

              return (
                <div
                  ref={provided.innerRef}
                  {...provided.draggableProps}
                  {...provided.dragHandleProps}
                  className={draggingBackground}
                >
                  {renderListItem(item, idx)}
                </div>
              )
            }}
          </Draggable>
        )
      }

      // render list item as subheader
      if (item.subHeader) {
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

      // render standard list item
      return renderListItem(item, idx)
    })
  }

  function renderListItem(item, idx) {
    // render custom list item element
    if (item.el) {
      return item.el
    }

    let itemProps = {}
    if (item.url) {
      itemProps = {
        component: Link,
        to: absURL(item.url),
        button: true,
      }
    }

    let className = classes.listItem
    if (item.highlight) {
      className = classnames(classes.listItem, classes.highlightedItem)
    }

    return (
      <ListItem key={idx} {...itemProps} className={className}>
        {item.icon && <ListItemIcon>{item.icon}</ListItemIcon>}
        <ListItemText
          primary={item.title}
          secondary={item.subText}
          secondaryTypographyProps={{ className: classes.listItemSubtext }}
        />
        {item.secondaryAction && (
          <ListItemSecondaryAction>
            {item.secondaryAction}
          </ListItemSecondaryAction>
        )}
      </ListItem>
    )
  }
}

/*
 * !!IMPORTANT!!
 *
 * Any additional properties not specified here will be
 * passed to the <List /> component
 */
FlatList.propTypes = {
  headerNote: p.node, // text displayed at the top of the list
  emptyMessage: p.string, // text displayed if list is empty

  // list items to render
  items: p.arrayOf(
    p.oneOfType([
      p.shape({
        highlight: p.bool, // changes the list item background color
        title: p.node.isRequired, // primary list item text
        subText: p.node, // secondary list item text
        secondaryAction: p.element, // right-most action
        url: p.string, // renders as a link routing to url
        icon: p.element, // renders a list item icon (or avatar)
        id: p.string, // required for drag and drop
      }),
      p.shape({
        subHeader: p.string.isRequired,
      }),
      p.shape({
        el: p.element.isRequired, // renders custom element
      }),
    ]),
  ),

  /*
   * If specified, enables drag and drop. Cache
   * updates to maintain a proper user experience
   * are expected to be handled within the
   * component calling FlatList.
   *
   * onReorder(id, oldIndex, newIndex)
   */
  onReorder: p.func,
}

FlatList.defaultProps = {
  items: [],
  emptyMessage: 'No results',
}
