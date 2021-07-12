import React, { ReactNode, useState, ReactElement, forwardRef } from 'react'
import { isWidthUp } from '@material-ui/core/withWidth'

import Avatar from '@material-ui/core/Avatar'
import Card from '@material-ui/core/Card'
import CircularProgress from '@material-ui/core/CircularProgress'
import Grid from '@material-ui/core/Grid'
import IconButton from '@material-ui/core/IconButton'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import Typography from '@material-ui/core/Typography'

import LeftIcon from '@material-ui/icons/ChevronLeft'
import RightIcon from '@material-ui/icons/ChevronRight'
import useWidth from '../util/useWidth'

import { FavoriteIcon } from '../util/SetFavoriteButton'
import { ITEMS_PER_PAGE } from '../config'
import ListItemAvatar from '@material-ui/core/ListItemAvatar'
import { makeStyles } from '@material-ui/core'
import InfiniteScroll from 'react-infinite-scroll-component'
import Spinner from '../loading/components/Spinner'
import { CheckboxItemsProps } from './ControlledPaginatedList'
import AppLink, { AppLinkProps } from '../util/AppLink'
import statusStyles from '../util/statusStyles'
import { debug } from '../util/debug'

// gray boxes on load
// disable overflow
// can go to last page + one if loading & hasNextPage
// delete on details -> update list (cache, refetch?)
// - on details, don't have accesses to search param

const useStyles = makeStyles((theme) => ({
  infiniteScrollFooter: {
    display: 'flex',
    justifyContent: 'center',
    padding: '0.25em 0 0.25em 0',
  },
  itemAction: {
    paddingLeft: 14,
  },
  itemText: {
    wordBreak: 'break-word',
  },
  progress: {
    color: theme.palette.secondary.main,
    position: 'absolute',
  },
  favoriteIcon: {
    backgroundColor: 'transparent',
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
  ...statusStyles,
}))

function PageControls(props: {
  isLoading: boolean
  onNext?: () => void
  onBack?: () => void
}): JSX.Element {
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
    minHeight: (dense) => (dense ? 57 : 71),
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
function LoadingItem(props: { dense?: boolean }): JSX.Element {
  const classes = useLoadingStyles(props.dense)

  return (
    <ListItem className={classes.item} dense={props.dense}>
      <ListItemText className={classes.lineOne} />
      <ListItemText className={classes.lineTwo} />
      <ListItemText className={classes.lineThree} />
    </ListItem>
  )
}

export interface PaginatedListProps {
  // cardHeader will be displayed at the top of the card
  cardHeader?: ReactNode

  // header elements will be displayed at the top of the list.
  headerNote?: string // left-aligned
  headerAction?: JSX.Element // right-aligned

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
  status?: 'ok' | 'warn' | 'err'
}

export function PaginatedList(props: PaginatedListProps): JSX.Element {
  const {
    cardHeader,
    headerNote,
    headerAction,
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

  function handleNextPage(): void {
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
    let favIcon = null
    if (item.isFavorite) {
      favIcon = (
        <div className={classes.itemAction}>
          <Avatar className={classes.favoriteIcon}>
            <FavoriteIcon />
          </Avatar>
        </div>
      )
    }

    // get status style for left-most border color
    let itemClass = classes.noStatus
    switch (item.status) {
      case 'ok':
        itemClass = classes.statusOK
        break
      case 'warn':
        itemClass = classes.statusWarning
        break
      case 'err':
        itemClass = classes.statusError
        break
    }

    const AppLinkListItem = forwardRef<HTMLAnchorElement, AppLinkProps>(
      (props, ref) => (
        <li>
          <AppLink ref={ref} {...props} />
        </li>
      ),
    )
    AppLinkListItem.displayName = 'AppLinkListItem'

    // must be explicitly set when using, in accordance with TS definitions
    const urlProps = item.url && {
      component: AppLinkListItem,

      // NOTE button: false? not assignable to true
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      button: true as any,
      to: item.url,
    }

    return (
      <ListItem
        className={itemClass}
        dense={isWidthUp('md', width)}
        key={'list_' + idx}
        {...urlProps}
      >
        {item.icon && <ListItemAvatar>{item.icon}</ListItemAvatar>}
        <ListItemText
          className={classes.itemText}
          primary={item.title}
          secondary={item.subText}
        />
        {favIcon}
        {item.action && <div className={classes.itemAction}>{item.action}</div>}
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

  function renderList(): ReactElement {
    return (
      <List data-cy='apollo-list'>
        {(headerNote || headerAction) && (
          <ListItem>
            {headerNote && (
              <ListItemText
                className={classes.headerNote}
                disableTypography
                secondary={
                  <Typography color='textSecondary'>{headerNote}</Typography>
                }
              />
            )}
            {headerAction && (
              <ListItemSecondaryAction>{headerAction}</ListItemSecondaryAction>
            )}
          </ListItem>
        )}
        {renderListItems()}
      </List>
    )
  }

  function renderAsInfiniteScroll(): ReactElement {
    const len = items.length

    return (
      <InfiniteScroll
        hasMore={Boolean(loadMore)}
        next={
          loadMore ||
          (() => {
            debug('next callback missing from InfiniteScroll')
          })
        }
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
}
