import React, {
  ReactNode,
  useState,
  ReactElement,
  Dispatch,
  SetStateAction,
} from 'react'
import { isWidthUp } from '@material-ui/core/withWidth'
import { useDispatch, useSelector } from 'react-redux'

import Avatar from '@material-ui/core/Avatar'
import FavoriteIcon from '@material-ui/icons/Star'
import Card from '@material-ui/core/Card'
import CircularProgress from '@material-ui/core/CircularProgress'
import Grid from '@material-ui/core/Grid'
import IconButton from '@material-ui/core/IconButton'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Typography from '@material-ui/core/Typography'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'

import LeftIcon from '@material-ui/icons/ChevronLeft'
import RightIcon from '@material-ui/icons/ChevronRight'
import { Link } from 'react-router-dom'
import useWidth from '../util/useWidth'

import { ITEMS_PER_PAGE } from '../config'
import { absURLSelector } from '../selectors/url'
import { setCheckedItems as _setCheckedItems } from '../actions'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import { Checkbox, makeStyles } from '@material-ui/core'
import InfiniteScroll from 'react-infinite-scroll-component'
import Spinner from '../loading/components/Spinner'
import ListControls, { CheckboxActions } from './ListControls'

// gray boxes on load
// disable overflow
// can go to last page + one if loading & hasNextPage
// delete on details -> update list (cache, refetch?)
// - on details, don't have accesses to search param

const useStyles = makeStyles(theme => ({
  infiniteScrollFooter: {
    display: 'flex',
    justifyContent: 'center',
    padding: '0.25em 0 0.25em 0',
  },
  progress: {
    color: theme.palette.secondary.main,
    position: 'absolute',
  },
  favoriteIcon: {
    backgroundColor: 'transparent',
    color: 'grey',
  },
  listHeader: {
    fontStyle: 'italic',
  },
  controls: {
    [theme.breakpoints.down('sm')]: {
      '&:not(:first-child)': {
        marginBottom: '4.5em',
        paddingBottom: '1em',
      },
    },
  },
}))

export interface PaginatedListProps {
  // cardHeader will be displayed at the top of the card
  cardHeader?: ReactNode

  // listHeader will be displayed at the top of the list
  listHeader?: ReactNode

  items: PaginatedListItemProps[]

  // renders checkboxes for ListControls actions next to each list item
  // NOTE: this will replace any icons set on each item with a checkbox
  withCheckboxes?: boolean

  itemsPerPage: number

  isLoading?: boolean
  loadMore?: any

  // disables the placeholder display during loading
  noPlaceholder?: boolean

  // provide a custom message to display if there are no results
  emptyMessage?: string

  // if set, loadMore will be called when the user
  // scrolls to the bottom of the list. appends list
  // items to the list rather than rendering a new page
  infiniteScroll?: boolean
}

export interface PaginatedListItemProps {
  id: string
  url?: string
  title: string
  subText?: string
  isFavorite?: boolean
  icon?: ReactElement // renders a list item icon (or avatar)
  action?: ReactNode
  className?: string
}

export function PaginatedList(props: PaginatedListProps) {
  const {
    cardHeader,
    listHeader,
    items = [],
    itemsPerPage = ITEMS_PER_PAGE,
    infiniteScroll,
    loadMore,
    emptyMessage = 'No results',
    noPlaceholder,
    withCheckboxes,
  } = props

  const classes = useStyles()
  const absURL = useSelector(absURLSelector)

  const dispatch = useDispatch()
  // @ts-ignore
  const checkedItems = useSelector(state => state.list.checkedItems)
  const setCheckedItems = (array: Array<any>) =>
    dispatch(_setCheckedItems(array))

  const [page, setPage] = useState(0)

  const pageCount = Math.ceil(items.length / itemsPerPage)
  const width = useWidth()

  // isLoading returns true if the parent says we are, or
  // we are currently on an incomplete page and `loadMore` is available.
  const isLoading = (() => {
    if (props.isLoading) return true

    // We are on a future/incomplete page and loadMore is true
    const itemCount = items.length
    if ((page + 1) * itemsPerPage > itemCount && loadMore) return true

    return false
  })()

  const hasNextPage = (() => {
    const nextPage = page + 1

    // Check that we have at least 1 item already for the next page
    if (nextPage < pageCount) return true

    // If we're on the last page, not already loading, and can load more
    if (nextPage === pageCount && !isLoading && loadMore) {
      return true
    }

    return false
  })()

  function handleNextPage() {
    const nextPage = page + 1
    setPage(nextPage)

    // If we're on a not-fully-loaded page, or the last page when > the first page
    if (
      (nextPage >= pageCount || (nextPage > 1 && nextPage + 1 === pageCount)) &&
      loadMore
    )
      loadMore()
  }

  function renderNoResults() {
    return (
      <ListItem>
        <ListItemText
          disableTypography
          secondary={<Typography variant='caption'>{emptyMessage}</Typography>}
        />
      </ListItem>
    )
  }

  function renderItem(item: PaginatedListItemProps, idx: number) {
    let favIcon = <ListItemSecondaryAction />

    if (item.isFavorite) {
      favIcon = (
        <ListItemSecondaryAction>
          <Avatar className={classes.favoriteIcon}>
            <FavoriteIcon data-cy='fav-icon' />
          </Avatar>
        </ListItemSecondaryAction>
      )
    }

    // must be explicitly set when using, in accordance with TS definitions
    const urlProps = item.url && {
      component: Link,
      button: true as any,
      to: absURL(item.url),
    }

    let checkbox = null
    if (withCheckboxes) {
      const checked = checkedItems.includes(item.id)
      // TODO: custom props, e.g. disabled for closed alerts
      checkbox = (
        <Checkbox
          checked={checked}
          data-cy={'item-' + item.id}
          onClick={e => {
            e.stopPropagation()
            e.preventDefault()

            if (checked) {
              const idx = checkedItems.indexOf(item.id)
              const newItems = checkedItems.slice()
              newItems.splice(idx, 1)
              setCheckedItems(newItems)
            } else {
              setCheckedItems([...checkedItems, item.id])
            }
          }}
        />
      )
    }

    return (
      <ListItem
        className={item.className}
        dense={isWidthUp('md', width)}
        key={'list_' + idx}
        {...urlProps}
      >
        {checkbox && <ListItemIcon>{checkbox}</ListItemIcon>}
        {item.icon && !checkbox && <ListItemIcon>{item.icon}</ListItemIcon>}
        <ListItemText primary={item.title} secondary={item.subText} />
        {favIcon}
        {item.action && (
          <ListItemSecondaryAction>{item.action}</ListItemSecondaryAction>
        )}
      </ListItem>
    )
  }

  function renderListItems() {
    if (pageCount === 0 && !isLoading) return renderNoResults()

    let renderedItems: any = items
    if (!infiniteScroll) {
      renderedItems = items.slice(
        page * itemsPerPage,
        (page + 1) * itemsPerPage,
      )
    }
    renderedItems = renderedItems.map(renderItem)

    // Display full list when loading
    if (!noPlaceholder) {
      while (isLoading && renderedItems.length < itemsPerPage) {
        renderedItems.push(
          <LoadingItem
            dense={isWidthUp('md', width)}
            key={'list_' + renderedItems.length}
          />,
        )
      }
    }

    return renderedItems
  }

  let onBack = page > 0 ? () => setPage(page - 1) : undefined
  let onNext = hasNextPage ? handleNextPage : undefined

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Card>
          {cardHeader}
          {infiniteScroll ? renderAsInfiniteScroll() : renderList()}
        </Card>
      </Grid>
      {!infiniteScroll && (
        <PageControls onBack={onBack} onNext={onNext} isLoading={isLoading} />
      )}
    </Grid>
  )

  function renderList() {
    return (
      <List data-cy='apollo-list'>
        {listHeader && (
          <ListItem>
            <ListItemText
              className={classes.listHeader}
              disableTypography
              secondary={
                <Typography color='textSecondary'>{listHeader}</Typography>
              }
            />
          </ListItem>
        )}
        {renderListItems()}
      </List>
    )
  }

  function renderAsInfiniteScroll() {
    const len = items.length

    return (
      <InfiniteScroll
        scrollableTarget='content'
        next={loadMore}
        hasMore={Boolean(loadMore)}
        endMessage={
          len === 0 ? null : (
            <Typography
              className={classes.infiniteScrollFooter}
              color='textSecondary'
              variant='body2'
            >
              Displaying all results.
            </Typography>
          )
        }
        loader={
          <div className={classes.infiniteScrollFooter}>
            <Spinner text='Loading...' />
          </div>
        }
        dataLength={len}
      >
        {renderList()}
      </InfiniteScroll>
    )
  }
}

function PageControls(props: {
  isLoading: boolean
  onNext?: () => void
  onBack?: () => void
}) {
  const classes = useStyles()
  const { isLoading, onBack, onNext } = props

  return (
    <Grid
      item // item within main render grid
      xs={12}
      container // container for control items
      spacing={1}
      justify='flex-end'
      alignItems='center'
      className={classes.controls}
    >
      <Grid item>
        <IconButton
          title='back page'
          data-cy='back-button'
          disabled={!onBack}
          onClick={() => {
            onBack && onBack()
            window.scrollTo(0, 0)
          }}
        >
          <LeftIcon />
        </IconButton>
      </Grid>
      <Grid item>
        <IconButton
          title='next page'
          data-cy='next-button'
          disabled={!onNext}
          onClick={() => {
            onNext && onNext()
            window.scrollTo(0, 0)
          }}
        >
          {isLoading && !onNext && (
            <CircularProgress
              color='secondary'
              size={24}
              className={classes.progress}
            />
          )}
          <RightIcon />
        </IconButton>
      </Grid>
    </Grid>
  )
}

const loadingStyle = {
  color: 'lightgrey',
  background: 'lightgrey',
  height: '10.3333px',
}

const useLoadingStyles = makeStyles({
  item: {
    display: 'block',
    minHeight: dense => (dense ? 57 : 71),
  },
  lineOne: {
    ...loadingStyle,
    width: '50%',
  },
  lineTwo: {
    ...loadingStyle,
    width: '35%',
    margin: '5px 0 5px 0',
  },
  lineThree: {
    ...loadingStyle,
    width: '65%',
  },
})

// LoadingItem is used as a placeholder for loading content
function LoadingItem(props: { dense?: boolean }) {
  const classes = useLoadingStyles(props.dense)

  return (
    <ListItem className={classes.item} dense={props.dense}>
      <ListItemText className={classes.lineOne} />
      <ListItemText className={classes.lineTwo} />
      <ListItemText className={classes.lineThree} />
    </ListItem>
  )
}
