import React, {
  useEffect,
  useRef,
  useState,
  MouseEvent,
  ReactNode,
} from 'react'
import ButtonBase from '@mui/material/ButtonBase'
import IconButton from '@mui/material/IconButton'
import List, { ListProps } from '@mui/material/List'
import MUIListItem, { ListItemProps } from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import Typography from '@mui/material/Typography'
import ListSubheader from '@mui/material/ListSubheader'
import makeStyles from '@mui/styles/makeStyles'
import { CSSTransition, TransitionGroup } from 'react-transition-group'
import {
  Alert,
  AlertTitle,
  ListItemButton,
  ListItemIcon,
  Collapse,
} from '@mui/material'
import EditIcon from '@mui/icons-material/Edit'
import DoneIcon from '@mui/icons-material/Done'
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
import { Notice, toSeverity } from '../details/Notices'
import FlatListItem from './FlatListItem'
import { DraggableListItem, getAnnouncements } from './DraggableListItem'
import { ExpandLess, ExpandMore } from '@mui/icons-material'

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
export interface FlatListItem extends ListItemProps {
  title?: string
  highlight?: boolean
  subText?: JSX.Element | string
  icon?: JSX.Element | null
  section?: string | number
  secondaryAction?: JSX.Element | null
  url?: string
  id?: string // required for drag and drop functionality
  scrollIntoView?: boolean
  'data-cy'?: string
  disabled?: boolean
}

export interface SectionTitle {
  title: string
  icon?: JSX.Element | null
  subText?: JSX.Element | string
}

export type FlatListListItem = FlatListSub | FlatListItem | FlatListNotice

export interface FlatListProps extends ListProps {
  items: FlatListListItem[]

  // sectition titles for collaspable sections
  sections?: SectionTitle[]

  // header elements will be displayed at the top of the list.
  headerNote?: JSX.Element | string | ReactNode // left-aligned
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

  // will render items in collaspable sections in list
  collapsable?: boolean

  // renders an edit button that hides the options buttons until toggled on
  toggleDnD?: boolean
}

export default function FlatList({
  onReorder,
  emptyMessage,
  headerNote,
  headerAction,
  items,
  inset,
  sections,
  transition,
  collapsable,
  toggleDnD,
  ...listProps
}: FlatListProps): JSX.Element {
  const classes = useStyles()

  // collapsable sections state
  const [openSections, setOpenSections] = useState<string[]>(
    sections && sections.length ? [sections[0].title] : [],
  )

  useEffect(() => {
    const sectionArr = sections?.map((section) => section.title)
    // update openSections if there are new sections
    if (
      openSections.length &&
      sectionArr?.length &&
      !sectionArr?.find((section: string) => section === openSections[0])
    ) {
      setOpenSections([sectionArr[0]])
    }
  }, [sections])

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

  useEffect(() => {
    setDndItems(items.map((i, idx) => (i.id ? i.id : idx.toString())))
  }, [items])

  const [dragging, setDragging] = useState(false)
  const [draggable, setDraggable] = useState(false)
  const isFirstAnnouncement = useRef(false)
  const announcements = getAnnouncements(dndItems, isFirstAnnouncement)
  function handleDragStart(): void {
    if (!isFirstAnnouncement.current) {
      isFirstAnnouncement.current = true
    }
    setDragging(true)
  }
  function handleDragEnd(e: DragEndEvent): void {
    setDragging(false)
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
            severity={toSeverity(item.type)}
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
        severity={toSeverity(item.type)}
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
            showOptions={toggleDnD ? draggable : true}
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
            draggable={toggleDnD ? draggable : true}
          />
        )
      }

      return (
        <FlatListItem
          key={`${idx}-${item.id}`}
          index={idx}
          item={item}
          showOptions={toggleDnD ? draggable : true}
        />
      )
    })
  }

  function renderCollapsableItems(): JSX.Element[] | undefined {
    const toggleSection = (section: string): void => {
      if (openSections?.includes(section)) {
        setOpenSections(
          openSections.filter((openSection) => openSection !== section),
        )
      } else {
        setOpenSections([...openSections, section])
      }
    }
    return sections?.map((section, idx) => {
      const open = openSections?.includes(section.title)
      return (
        <React.Fragment key={idx}>
          <ListItemButton onClick={() => toggleSection(section.title)}>
            {section.icon && <ListItemIcon>{section.icon}</ListItemIcon>}
            <ListItemText primary={section.title} secondary={section.subText} />
            {open ? <ExpandLess /> : <ExpandMore />}
          </ListItemButton>
          <Collapse in={open} unmountOnExit>
            <List>
              {items
                .filter((item: FlatListItem) => item.section === section.title)
                .map((item, idx) => {
                  return (
                    <FlatListItem
                      index={idx}
                      key={idx}
                      item={item}
                      showOptions={toggleDnD ? draggable : true}
                    />
                  )
                })}
            </List>
          </Collapse>
        </React.Fragment>
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

    const renderListItems = ():
      | (JSX.Element | undefined)[]
      | JSX.Element
      | JSX.Element[]
      | undefined => {
      if (transition) return renderTransitions()
      if (collapsable) return renderCollapsableItems()
      return renderItems()
    }

    return (
      <List {...listProps} sx={sx}>
        {(headerNote || headerAction || onReorder) && (
          <MUIListItem>
            {toggleDnD && (
              <IconButton
                onClick={() => setDraggable(!draggable)}
                disabled={draggable && dragging}
                sx={{
                  marginRight: (t) => t.spacing(2),
                  textOverflow: 'wrap',
                }}
                aria-label='Toggle Drag and Drop'
              >
                {draggable ? <DoneIcon /> : <EditIcon />}
              </IconButton>
            )}
            {headerNote && (
              <ListItemText
                disableTypography
                secondary={
                  <Typography color='textSecondary'>{headerNote}</Typography>
                }
                sx={{ fontStyle: 'italic', pr: 2 }}
              />
            )}
            {headerAction && <div>{headerAction}</div>}
          </MUIListItem>
        )}
        {!items.length && renderEmptyMessage()}
        {renderListItems()}
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
