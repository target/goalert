import React, { ReactNode, ReactElement, forwardRef, useContext } from 'react'
import Avatar from '@material-ui/core/Avatar'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Typography from '@material-ui/core/Typography'
import ListItemAvatar from '@material-ui/core/ListItemAvatar'
import { makeStyles } from '@material-ui/core'
import { useIsWidthDown } from '../util/useWidth'
import { FavoriteIcon } from '../util/SetFavoriteButton'
import { ITEMS_PER_PAGE } from '../config'
import InfiniteScroll from 'react-infinite-scroll-component'
import Spinner from '../loading/components/Spinner'
import { CheckboxItemsProps } from './ControlledPaginatedList'
import AppLink, { AppLinkProps } from '../util/AppLink'
import statusStyles from '../util/statusStyles'
import { debug } from '../util/debug'
import { PageControlsContext } from './PageControls'

// gray boxes on load
// disable overflow
// can go to last page + one if loading & hasNextPage
// delete on details -> update list (cache, refetch?)
// - on details, don't have accesses to search param

const useStyles = makeStyles(() => ({
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
  favoriteIcon: {
    backgroundColor: 'transparent',
  },
  ...statusStyles,
}))

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
  items: PaginatedListItemProps[] | CheckboxItemsProps[]
  itemsPerPage?: number

  pageCount?: number

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
    items = [],
    itemsPerPage = ITEMS_PER_PAGE,
    pageCount,
    infiniteScroll,
    isLoading,
    loadMore,
    emptyMessage = 'No results',
    noPlaceholder,
  } = props

  const { page } = useContext(PageControlsContext)

  const classes = useStyles()

  const fullScreen = useIsWidthDown('md')

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
        dense={!fullScreen}
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
            dense={!fullScreen}
            key={'list_' + renderedItems.length}
          />,
        )
      }
    }

    return renderedItems
  }

  function renderList(): ReactElement {
    return <List data-cy='apollo-list'>{renderListItems()}</List>
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
    <React.Fragment>
      {infiniteScroll ? renderAsInfiniteScroll() : renderList()}
    </React.Fragment>
  )
}
