import React from 'react'
import p from 'prop-types'
import withStyles from '@material-ui/core/styles/withStyles'
import withWidth, { isWidthUp } from '@material-ui/core/withWidth/index'

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
import { connect } from 'react-redux'

import { ITEMS_PER_PAGE } from '../config'
import { absURLSelector } from '../selectors/url'
import ListItemIcon from '@material-ui/core/ListItemIcon'

// gray boxes on load
// disable overflow
// can go to last page + one if loading & hasNextPage
// delete on details -> update list (cache, refetch?)
// - on details, don't have accesses to search param

const styles = theme => ({
  progress: {
    color: theme.palette.secondary['500'],
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
})

@withStyles(styles)
class PaginationControls extends React.PureComponent {
  static propTypes = {
    isLoading: p.bool,
    onNext: p.func,
    onBack: p.func,
  }

  render() {
    const { classes, isLoading, onBack, onNext } = this.props

    return (
      <React.Fragment>
        <Grid container justify='flex-end' className={classes.controls}>
          <Grid item>
            <IconButton
              title='back page'
              data-cy='back-button'
              disabled={!onBack}
              onClick={() => {
                onBack()
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
                onNext()
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
      </React.Fragment>
    )
  }
}

const loadingStyle = {
  color: 'lightgrey',
  background: 'lightgrey',
  height: '10.3333px',
}

// LoadingItem is used as a placeholder for loading content
class LoadingItem extends React.PureComponent {
  render() {
    const { dense } = this.props
    let minHeight = 71
    if (dense) {
      minHeight = 57
    }
    return (
      <ListItem dense={dense} style={{ display: 'block', minHeight }}>
        <ListItemText style={{ ...loadingStyle, width: '50%' }} />
        <ListItemText
          style={{ ...loadingStyle, width: '35%', margin: '5px 0 5px 0' }}
        />
        <ListItemText style={{ ...loadingStyle, width: '65%' }} />
      </ListItem>
    )
  }
}

const mapStateToProps = state => {
  return {
    absURL: absURLSelector(state),
  }
}

@withWidth()
@withStyles(styles)
@connect(mapStateToProps)
export class PaginatedList extends React.PureComponent {
  static propTypes = {
    // headerNote will be displayed at the top of the list.
    headerNote: p.oneOfType([p.string, p.element]),

    items: p.arrayOf(
      p.shape({
        url: p.string,
        title: p.string.isRequired,
        subText: p.string,
        isFavorite: p.bool,
        icon: p.element, // renders a list item icon (or avatar)
        action: p.element,
      }),
    ),

    isLoading: p.bool,
    loadMore: p.func,

    // disable placeholder display during loading
    noPlaceholder: p.bool,

    // provide a message to display if there are no results
    emptyMessage: p.string,
  }

  static defaultProps = {
    emptyMessage: 'No results',
  }

  state = {
    page: 0,
  }

  pageCount = () => Math.ceil((this.props.items || []).length / ITEMS_PER_PAGE)

  // isLoading returns true if the parent says we are, or
  // we are currently on an incomplete page and `loadMore` is available.
  isLoading() {
    if (this.props.isLoading) return true

    // We are on a future/incomplete page and loadMore is true
    const itemCount = (this.props.items || []).length
    if (
      (this.state.page + 1) * ITEMS_PER_PAGE > itemCount &&
      this.props.loadMore
    )
      return true

    return false
  }

  hasNextPage() {
    const nextPage = this.state.page + 1
    // Check that we have at least 1 item already for the next page
    if (nextPage < this.pageCount()) return true

    // If we're on the last page, not already loading, and can load more
    if (
      nextPage === this.pageCount() &&
      !this.isLoading() &&
      this.props.loadMore
    ) {
      return true
    }

    return false
  }

  onNextPage = () => {
    const nextPage = this.state.page + 1
    this.setState({ page: nextPage })

    // If we're on a not-fully-loaded page, or the last page when > the first page
    if (
      (nextPage >= this.pageCount() ||
        (nextPage > 1 && nextPage + 1 === this.pageCount())) &&
      this.props.loadMore
    )
      this.props.loadMore(ITEMS_PER_PAGE * 2)
  }

  renderPaginationControls() {
    let onBack = null
    let onNext = null

    if (this.state.page > 0)
      onBack = () => this.setState({ page: this.state.page - 1 })
    if (this.hasNextPage()) onNext = this.onNextPage

    return (
      <PaginationControls
        onBack={onBack}
        onNext={onNext}
        isLoading={this.isLoading()}
      />
    )
  }

  renderNoResults() {
    return (
      <ListItem>
        <ListItemText
          disableTypography
          secondary={
            <Typography variant='caption'>{this.props.emptyMessage}</Typography>
          }
        />
      </ListItem>
    )
  }

  renderItem = (item, idx) => {
    const { classes, width, absURL } = this.props

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

    return (
      <ListItem
        dense={isWidthUp('md', width)}
        key={'list_' + idx}
        component={item.url ? Link : null}
        to={absURL(item.url)}
        button={Boolean(item.url)}
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

  renderListItems() {
    if (this.pageCount() === 0 && !this.props.isLoading)
      return this.renderNoResults()

    const { page } = this.state
    const { width, noPlaceholder } = this.props

    const items = (this.props.items || [])
      .slice(page * ITEMS_PER_PAGE, (page + 1) * ITEMS_PER_PAGE)
      .map(this.renderItem)

    // Display full list when loading
    if (!noPlaceholder) {
      while (this.isLoading() && items.length < ITEMS_PER_PAGE) {
        items.push(
          <LoadingItem
            dense={isWidthUp('md', width)}
            key={'list_' + items.length}
          />,
        )
      }
    }

    return items
  }

  render() {
    const { headerNote, classes } = this.props
    return (
      <React.Fragment>
        <Grid item xs={12}>
          <Card>
            <List data-cy='apollo-list'>
              {headerNote && (
                <ListItem>
                  <ListItemText
                    className={classes.headerNote}
                    disableTypography
                    secondary={
                      <Typography color='textSecondary'>
                        {headerNote}
                      </Typography>
                    }
                  />
                </ListItem>
              )}
              {this.renderListItems()}
            </List>
          </Card>
        </Grid>
        {this.renderPaginationControls()}
      </React.Fragment>
    )
  }
}
