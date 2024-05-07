import React, { ReactNode, ReactElement } from 'react'
import Avatar from '@mui/material/Avatar'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import Typography from '@mui/material/Typography'
import ListItemAvatar from '@mui/material/ListItemAvatar'
import makeStyles from '@mui/styles/makeStyles'
import { Skeleton } from '@mui/material'
import InfiniteScroll from 'react-infinite-scroll-component'
import { useIsWidthDown } from '../util/useWidth'
import { FavoriteIcon } from '../util/SetFavoriteButton'
import { ITEMS_PER_PAGE } from '../config'
import Spinner from '../loading/components/Spinner'
import { CheckboxItemsProps } from './ControlledPaginatedList'
import { debug } from '../util/debug'
import useStatusColors from '../theme/useStatusColors'

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
}))

function LoadingItem(props: { dense?: boolean }): JSX.Element {
  return (
    <ListItem dense={props.dense}>
      <Skeleton variant='rectangular' animation='wave' width='100%'>
        <ListItemText primary='.' secondary='.' />
      </Skeleton>
    </ListItem>
  )
}

export interface PaginatedListProps {
  items: PaginatedListItemProps[] | CheckboxItemsProps[]
  itemsPerPage?: number

  page?: number

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
  subText?: ReactNode
  isFavorite?: boolean
  icon?: ReactElement // renders a list item icon (or avatar)
  action?: ReactNode
  status?: 'ok' | 'warn' | 'err'
  onClick?: () => void
  selected?: boolean
}

export function PaginatedList(props: PaginatedListProps): JSX.Element {
  const {
    items = [],
    itemsPerPage = ITEMS_PER_PAGE,
    page,
    infiniteScroll,
    isLoading,
    loadMore,
    emptyMessage = 'No results',
    noPlaceholder,
  } = props

  const classes = useStyles()
  const statusColors = useStatusColors()
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

    const borderColor = (s?: string): string => {
      switch (s) {
        case 'ok':
        case 'warn':
        case 'err':
          return statusColors[s]

        default:
          return 'transparent'
      }
    }

    // must be explicitly set when using, in accordance with TS definitions
    const urlProps = item.url && {
      component: 'li',

      // NOTE: needed for error: button: false? not assignable to type 'true'
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      button: true as any,
      to: item.url,
    }
    const onClickProps = item.onClick && {
      onClick: item.onClick,

      // NOTE: needed for error: button: false? not assignable to type 'true'
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      button: true as any,
    }

    return (
      <ListItem
        sx={{
          borderLeft: `3px solid ${borderColor(item.status)}`,
        }}
        dense={!fullScreen}
        key={'list_' + idx}
        selected={item.selected}
        {...urlProps}
        {...onClickProps}
      >
        {item.icon && <ListItemAvatar>{item.icon}</ListItemAvatar>}
        <ListItemText
          className={classes.itemText}
          primary={item.title}
          secondary={item.subText}
          secondaryTypographyProps={{
            component: typeof item.subText !== 'string' ? 'div' : 'p',
          }}
        />
        {favIcon}
        {item.action && <div className={classes.itemAction}>{item.action}</div>}
      </ListItem>
    )
  }

  function renderListItems(): ReactElement | ReactElement[] {
    if (items.length === 0 && !isLoading) return renderNoResults()

    let newItems: Array<PaginatedListItemProps> = items.slice()
    if (!infiniteScroll && page !== undefined) {
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
    const dataCy = isLoading ? 'paginated-list-loading' : 'paginated-list'
    return <List data-cy={dataCy}>{renderListItems()}</List>
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
