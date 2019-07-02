import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import AppBar from '@material-ui/core/AppBar'
import Hidden from '@material-ui/core/Hidden'
import IconButton from '@material-ui/core/IconButton'
import Slide from '@material-ui/core/Slide'
import TextField from '@material-ui/core/TextField'
import Toolbar from '@material-ui/core/Toolbar'
import withStyles from '@material-ui/core/styles/withStyles'
import { Close as CloseIcon, Search as SearchIcon } from '@material-ui/icons'
import { styles } from '../styles/materialStyles'
import { connect } from 'react-redux'
import { searchSelector } from '../selectors/url'
import { setURLParam } from '../actions/main'
import { debounce } from 'lodash-es'
import { DEBOUNCE_DELAY } from '../config'

const mapDispatchToProps = dispatch => {
  return {
    setSearch: debounce(
      value => dispatch(setURLParam('search', value)),
      DEBOUNCE_DELAY,
    ),
  }
}
const mapStateToProps = state => {
  return {
    search: searchSelector(state),
  }
}

/*
 * Renders a search bar that will fix to the top right of the screen (in the app bar)
 *
 * On a mobile device the the search icon will be present, and when tapped
 * a new appbar will display that contains a search field to use.
 *
 * On a larger screen, the field will always be present to use in the app bar.
 *
 * Uncontrolled component with a key. If the component detects that the search
 * URL parameter has been reset, it will reset the search param's state as well.
 * i.e. this component's location key changes and will remount a new version of itself.
 *
 * A search function is provided for components that need tracking of the search
 * in Redux (for cache updates, queries, etc).
 */
@withStyles(styles)
@connect(
  mapStateToProps,
  mapDispatchToProps,
)
export default class Search extends Component {
  static propTypes = {
    setSearch: p.func.isRequired, // used for redux updates
    search: p.string.isRequired, // initial value of search param, if given
  }

  state = {
    showMobile: false,
  }

  renderTextField = extraProps => {
    const { classes, search } = this.props

    return (
      <TextField
        key={search}
        autoFocus
        InputProps={{
          disableUnderline: true,
          classes: {
            input: classes.searchFieldBox,
          },
        }}
        placeholder='Search'
        onChange={e => this.props.setSearch(e.target.value)}
        defaultValue={search}
        {...extraProps}
      />
    )
  }

  renderMobileSearch() {
    return (
      <Hidden mdUp>
        <IconButton
          key='search-icon'
          color='inherit'
          aria-label='Search'
          data-cy='open-search'
          onClick={() => this.setState({ showMobile: true })}
        >
          <SearchIcon />
        </IconButton>
        <Slide
          key='search-field'
          in={this.state.showMobile}
          direction='down'
          mountOnEnter
          unmountOnExit
          style={{
            zIndex: 9001,
          }}
        >
          <AppBar>
            <Toolbar>
              <IconButton
                color='inherit'
                onClick={() => this.setState({ showMobile: false })}
                aria-label='Cancel'
                data-cy='close-search'
              >
                <CloseIcon />
              </IconButton>
              {this.renderTextField({ style: { flex: 1 } })}
            </Toolbar>
          </AppBar>
        </Slide>
      </Hidden>
    )
  }

  renderDesktopSearch() {
    return <Hidden smDown>{this.renderTextField()}</Hidden>
  }

  render() {
    return (
      <React.Fragment>
        {this.renderDesktopSearch()}
        {this.renderMobileSearch()}
      </React.Fragment>
    )
  }
}
