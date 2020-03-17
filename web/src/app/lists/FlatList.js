import React from 'react'
import p from 'prop-types'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import ListItemText from '@material-ui/core/ListItemText'
import Typography from '@material-ui/core/Typography'
import { DragDropContext, Droppable, Draggable } from 'react-beautiful-dnd'
import withStyles from '@material-ui/core/styles/withStyles'
import ListSubheader from '@material-ui/core/ListSubheader'
import { AppLink } from '../util/AppLink'

const styles = {
  background: { backgroundColor: 'white' },
  highlightedItem: {
    borderLeft: '6px solid #93ed94',
    background: '#defadf',
  },
  participantDragging: {
    backgroundColor: '#ebebeb',
  },
}

@withStyles(styles)
export default class FlatList extends React.PureComponent {
  static propTypes = {
    // headerNote will be displayed at the top of the list.
    headerNote: p.node,

    // emptyMessage will be displayed if there are no items in the list.
    emptyMessage: p.string,

    items: p.arrayOf(
      p.oneOfType([
        p.shape({
          highlight: p.bool,
          title: p.node.isRequired,
          subText: p.node,
          secondaryAction: p.element,
          url: p.string,
          icon: p.element, // renders a list item icon (or avatar)
          id: p.string, // required for drag and drop
        }),
        p.shape({
          subHeader: p.node.isRequired,
        }),
      ]),
    ),

    // indent text of each list item if no icon is present
    inset: p.bool,

    // If specified, enables drag and drop
    //
    // onReorder(id, oldIndex, newIndex)
    onReorder: p.func,
  }

  static defaultProps = {
    items: [],
    emptyMessage: 'No results',
  }

  handleDragStart = () => {
    // adds a little vibration if the browser supports it
    if (window.navigator.vibrate) {
      window.navigator.vibrate(100)
    }
  }

  handleDragEnd = result => {
    this.props.onReorder(
      // result.draggableId, : removed this as per new reorderList function
      result.source.index,
      result.destination.index,
    )
  }

  renderItem(item, idx) {
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
        style={{ width: '100%' }}
        className={item.highlight ? this.props.classes.highlightedItem : null}
      >
        {item.icon && <ListItemIcon>{item.icon}</ListItemIcon>}
        <ListItemText
          primary={item.title}
          secondary={item.subText}
          secondaryTypographyProps={{ style: { whiteSpace: 'pre-line' } }}
          inset={this.props.inset && !item.icon}
        />
        {item.secondaryAction && (
          <ListItemSecondaryAction>
            {item.secondaryAction}
          </ListItemSecondaryAction>
        )}
      </ListItem>
    )
  }

  renderItems() {
    if (!this.props.items.length) {
      return (
        <ListItem>
          <ListItemText
            disableTypography
            secondary={
              <Typography data-cy='list-empty-message' variant='caption'>
                {this.props.emptyMessage}
              </Typography>
            }
          />
        </ListItem>
      )
    }

    return this.props.items.map((item, idx) => {
      if (!this.props.onReorder) {
        if (item.subHeader) {
          return (
            <ListSubheader key={idx} className={this.props.classes.background}>
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
        return this.renderItem(item, idx)
      }
      return (
        <Draggable key={idx + item.id} draggableId={item.id} index={idx}>
          {(provided, snapshot) => {
            // light grey background while dragging non-active user
            const draggingBackground = snapshot.isDragging
              ? this.props.classes.participantDragging
              : null
            return (
              <div
                ref={provided.innerRef}
                {...provided.draggableProps}
                {...provided.dragHandleProps}
                className={draggingBackground}
              >
                {this.renderItem(item, idx)}
              </div>
            )
          }}
        </Draggable>
      )
    })
  }

  renderList() {
    const {
      dispatch,
      onReorder,
      classes,
      emptyMessage,
      headerNote,
      items,
      inset, // don't include in spread
      ...otherProps
    } = this.props

    return (
      <List {...otherProps}>
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
        {this.renderItems()}
      </List>
    )
  }

  renderDragAndDrop() {
    return (
      <DragDropContext
        onDragStart={this.handleDragStart}
        onDragEnd={this.handleDragEnd}
      >
        <Droppable droppableId='droppable'>
          {provided => (
            <div ref={provided.innerRef} {...provided.droppableProps}>
              {this.renderList()}
              {provided.placeholder}
            </div>
          )}
        </Droppable>
      </DragDropContext>
    )
  }

  render() {
    if (this.props.onReorder) {
      // Enable drag and drop
      return this.renderDragAndDrop()
    }
    return this.renderList()
  }
}
