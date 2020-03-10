import React, { ReactNode, useState, ReactElement } from 'react'
import { isWidthUp } from '@material-ui/core/withWidth'

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
import useWidth from '../util/useWidth'

import { ITEMS_PER_PAGE } from '../config'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import { makeStyles } from '@material-ui/core'
import InfiniteScroll from 'react-infinite-scroll-component'
import Spinner from '../loading/components/Spinner'
import { CheckboxItemsProps } from './ControlledPaginatedList'
import { AppLink } from '../util/AppLink'

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
  headerNote: {
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

  // headerNote will be displayed at the top of the list
  headerNote?: string

  items: PaginatedListItemProps[] | CheckboxItemsProps[]
  itemsPerPage?: number

  isLoading?: boolean
  loadMore?: (numberToLoad?: number) => void

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
  url?: string
  title: string
  subText?: string
  isFavorite?: boolean
  icon?: ReactElement // renders a list item icon (or avatar)
  action?: ReactNode
}

export function PaginatedList(props: PaginatedListProps) {
  const {
    cardHeader,
    headerNote,
    items = [],
    itemsPerPage = ITEMS_PER_PAGE,
    infiniteScroll,
    loadMore,
    emptyMessage = 'No results',
    noPlaceholder,
  } = props

  const classes = useStyles()

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
      loadMore(itemsPerPage * 2)
  }

  function renderNoResults(): ReactElement {
    return (
      <ListItem>
        <ListItemText
          disableTypography
          secondary={<Typography variant='caption'>{emptyMessage}</Typography>}
        />
      </ListItem>
    )
  }

  function renderItem(item: PaginatedListItemProps, idx: number): ReactElement {
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
      component: AppLink,
      button: true as any,
      to: item.url,
    }

    return (
      <ListItem
        dense={isWidthUp('md', width)}
        key={'list_' + idx}
        {...urlProps}
      >
        {item.icon && <ListItemIcon>{item.icon}</ListItemIcon>}
        <ListItemText primary={item.title} secondary={item.subText} />
        {favIcon}
        {item.action && (
          <ListItemSecondaryAction>{item.action}</ListItemSecondaryAction>
        )}
      </ListItem>
    )
  }

  function renderListItems(): ReactElement | ReactElement[] {
    if (pageCount === 0 && !isLoading) return renderNoResults()

    let newItems: Array<PaginatedListItemProps> = items.slice()
    if (!infiniteScroll) {
      newItems = items.slice(page * itemsPerPage, (page + 1) * itemsPerPage)
    }
    const renderedItems: ReactElement[] = newItems.map(renderItem)

    // Display full list when loading
    if (!noPlaceholder && isLoading) {
      while (renderedItems.length < itemsPerPage) {
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

  const onBack = page > 0 ? () => setPage(page - 1) : undefined
  const onNext = hasNextPage ? handleNextPage : undefined

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

  function renderList(): ReactElement {
    return (
      <List data-cy='apollo-list'>
        {headerNote && (
          <ListItem>
            <ListItemText
              className={classes.headerNote}
              disableTypography
              secondary={
                <Typography color='textSecondary'>{headerNote}</Typography>
              }
            />
          </ListItem>
        )}
        {renderListItems()}
      </List>
    )
  }

  function renderAsInfiniteScroll(): ReactElement {
    const len = items.length

    // explicitly set props to load more, if loader function present
    const loadProps: any = {}
    if (loadMore) {
      loadProps.hasMore = true
      loadProps.next = loadMore
    } else {
      loadProps.hasMore = false
    }

    return (
      <InfiniteScroll
        {...loadProps}
        scrollableTarget='content'
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
