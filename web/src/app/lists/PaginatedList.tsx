import React, { ReactNode, useState, ReactElement } from 'react'
import { isWidthUp } from '@material-ui/core/withWidth/index'
import { useSelector } from 'react-redux'

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
import ListItemIcon from '@material-ui/core/ListItemIcon'
import { makeStyles } from '@material-ui/core'

// gray boxes on load
// disable overflow
// can go to last page + one if loading & hasNextPage
// delete on details -> update list (cache, refetch?)
// - on details, don't have accesses to search param

const useStyles = makeStyles(theme => ({
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

// LoadingItem is used as a placeholder for loading content
function LoadingItem(props: { dense?: boolean }) {
  const minHeight = props.dense ? 57 : 71

  return (
    <ListItem dense={props.dense} style={{ display: 'block', minHeight }}>
      <ListItemText style={{ ...loadingStyle, width: '50%' }} />
      <ListItemText
        style={{ ...loadingStyle, width: '35%', margin: '5px 0 5px 0' }}
      />
      <ListItemText style={{ ...loadingStyle, width: '65%' }} />
    </ListItem>
  )
}

export interface PaginatedListProps {
  headerNote?: ReactNode

  items: PaginatedListItem[]

  isLoading?: boolean
  loadMore?: (numberToLoad: number) => void

  noPlaceholder?: boolean

  emptyMessage?: string
}

export interface PaginatedListItem {
  url?: string
  title: string
  subText?: string
  isFavorite?: boolean
  icon?: ReactElement
  action?: ReactNode
}

export function PaginatedList(props: PaginatedListProps) {
  const [page, setPage] = useState(0)
  const classes = useStyles()
  const absURL = useSelector(absURLSelector)

  const {
    items = [],
    loadMore,
    emptyMessage,
    noPlaceholder,
    headerNote,
  } = props
  const pageCount = Math.ceil(items.length / ITEMS_PER_PAGE)
  const width = useWidth()

  // isLoading returns true if the parent says we are, or
  // we are currently on an incomplete page and `loadMore` is available.
  const isLoading = (() => {
    if (props.isLoading) return true

    // We are on a future/incomplete page and loadMore is true
    const itemCount = items.length
    if ((page + 1) * ITEMS_PER_PAGE > itemCount && loadMore) return true

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
      loadMore(ITEMS_PER_PAGE * 2)
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

  function renderItem(item: PaginatedListItem, idx: number) {
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

    const extraProps = item.url && {
      component: Link,
      button: true as any,
      to: absURL(item.url),
    }
    return (
      <ListItem
        dense={isWidthUp('md', width)}
        key={'list_' + idx}
        {...extraProps}
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

  function renderListItems() {
    if (pageCount === 0 && !isLoading) return renderNoResults()

    const renderedItems = items
      .slice(page * ITEMS_PER_PAGE, (page + 1) * ITEMS_PER_PAGE)
      .map(renderItem)

    // Display full list when loading
    if (!noPlaceholder) {
      while (isLoading && items.length < ITEMS_PER_PAGE) {
        renderedItems.push(
          <LoadingItem
            dense={isWidthUp('md', width)}
            key={'list_' + items.length}
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
        </Card>
      </Grid>
      <PageControls onBack={onBack} onNext={onNext} isLoading={isLoading} />
    </Grid>
  )
}
