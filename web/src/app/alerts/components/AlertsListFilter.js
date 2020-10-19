import React, { Component } from 'react'
import p from 'prop-types'
import Button from '@material-ui/core/Button'
import IconButton from '@material-ui/core/IconButton'
import Popover from '@material-ui/core/Popover'
import FilterList from '@material-ui/icons/FilterList'
import Hidden from '@material-ui/core/Hidden'
import SwipeableDrawer from '@material-ui/core/SwipeableDrawer'
import Switch from '@material-ui/core/Switch'
import Grid from '@material-ui/core/Grid'
import { withStyles } from '@material-ui/core/styles'
import { styles as globalStyles } from '../../styles/materialStyles'
import Radio from '@material-ui/core/Radio'
import RadioGroup from '@material-ui/core/RadioGroup'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import FormControl from '@material-ui/core/FormControl'
import withWidth, { isWidthUp } from '@material-ui/core/withWidth'
import classnames from 'classnames'
import { connect } from 'react-redux'
import {
  resetAlertsFilters,
  setAlertsStatusFilter,
  setAlertsAllServicesFilter,
  setAlertsShowAsFullTimeFilter,
} from '../../actions'
import {
  alertAllServicesSelector,
  alertFilterSelector,
  alertShowAsFullTimeSelector,
} from '../../selectors'

const styles = (theme) => ({
  ...globalStyles(theme),
  drawer: {
    width: 'fit-content', // width placed on mobile drawer
  },
  grid: {
    margin: '0.5em', // margin in grid container
  },
  gridItem: {
    display: 'flex',
    alignItems: 'center', // aligns text with toggles
  },
  formControl: {
    width: '100%', // date pickers full width
  },
  popover: {
    width: '17em', // width placed on desktop popover
  },
})

const mapStateToProps = (state) => ({
  allServices: alertAllServicesSelector(state),
  filter: alertFilterSelector(state),
  showAsFullTime: alertShowAsFullTimeSelector(state),
})

const mapDispatchToProps = (dispatch) => ({
  resetAll: () => dispatch(resetAlertsFilters()), // don't reset search param
  setFilter: (value) => dispatch(setAlertsStatusFilter(value)),
  setAllServices: (value) => dispatch(setAlertsAllServicesFilter(value)),
  setShowAsFullTime: (value) =>
    dispatch(setAlertsShowAsFullTimeFilter(value)),
})

@withStyles(styles)
@withWidth()
@connect(mapStateToProps, mapDispatchToProps)
export default class AlertsListFilter extends Component {
  static propTypes = {
    serviceID: p.string,
    allServices: p.bool,
    filter: p.string,
    showAsFullTime: p.bool,
  }

  state = {
    show: false,
    anchorEl: null, // element in which filters form under
  }

  handleOpenFilters = (event) => {
    this.setState({
      anchorEl: event.currentTarget,
      show: true,
    })
  }

  handleCloseFilters = () => {
    this.setState({
      show: false,
    })
  }

  renderFilters = () => {
    const {
      allServices,
      classes,
      filter,
      serviceID: sid,
      showAsFullTime,
      width,
    } = this.props
    const {
      resetAll,
      setFilter,
      setAllServices,
      setShowAsFullTime,
    } = this.props

    // grabs class for width depending on breakpoints (md or higher uses popover width)
    const widthClass = isWidthUp('md', width) ? classes.popover : classes.drawer
    const gridClasses = classnames(classes.grid, widthClass)

    let favoritesFilter = null
    if (!sid) {
      favoritesFilter = (
        <FormControlLabel
          control={
            <Switch
              aria-label='Include All Services Toggle'
              data-cy='toggle-favorites'
              checked={allServices}
              onChange={() => setAllServices(!allServices)}
            />
          }
          label='Include All Services'
        />
      )
    }

    const content = (
      <Grid container spacing={2} className={gridClasses}>
        <Grid item xs={12} className={classes.gridItem}>
          <FormControl>
            {favoritesFilter}
            <FormControlLabel
              control={
                <Switch
                  aria-label='Show full timestamps toggle'
                  data-cy='toggle-full-time'
                  checked={showAsFullTime}
                  onChange={() => setShowAsFullTime(!showAsFullTime)}
                />
              }
              label='Show full timestamps'
            />
            <RadioGroup
              aria-label='Alert Status Filters'
              name='status-filters'
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
            >
              <FormControlLabel
                value='active'
                control={<Radio color='primary' />}
                label='Active'
              />
              <FormControlLabel
                value='unacknowledged'
                control={<Radio color='primary' />}
                label='Unacknowledged'
              />
              <FormControlLabel
                value='acknowledged'
                control={<Radio color='primary' />}
                label='Acknowledged'
              />
              <FormControlLabel
                value='closed'
                control={<Radio color='primary' />}
                label='Closed'
              />
              <FormControlLabel
                value='all'
                control={<Radio color='primary' />}
                label='All'
              />
            </RadioGroup>
          </FormControl>
        </Grid>
        <Grid item xs={12} className={classes.filterActions}>
          <Button onClick={resetAll}>Reset</Button>
          <Button onClick={this.handleCloseFilters}>Done</Button>
        </Grid>
      </Grid>
    )

    // renders a popover on desktop, and a swipeable drawer on mobile devices
    return (
      <React.Fragment>
        <Hidden smDown>
          <Popover
            anchorEl={() => this.state.anchorEl}
            open={!!this.state.anchorEl && this.state.show}
            onClose={this.handleCloseFilters}
            anchorOrigin={{
              vertical: 'bottom',
              horizontal: 'right',
            }}
            transformOrigin={{
              vertical: 'top',
              horizontal: 'right',
            }}
          >
            {content}
          </Popover>
        </Hidden>
        <Hidden mdUp>
          <SwipeableDrawer
            anchor='top'
            disableDiscovery
            disableSwipeToOpen
            open={this.state.show}
            onClose={this.handleCloseFilters}
            onOpen={this.handleOpenFilters}
          >
            {content}
          </SwipeableDrawer>
        </Hidden>
      </React.Fragment>
    )
  }

  /*
   * Finds the parent toolbar DOM node and appends the options
   * element to that node (after all the toolbar's children
   * are done being rendered)
   */
  render() {
    return (
      <React.Fragment>
        <IconButton
          aria-label='Filter Alerts'
          color='inherit'
          onClick={this.handleOpenFilters}
        >
          <FilterList />
        </IconButton>
        {this.renderFilters()}
      </React.Fragment>
    )
  }
}
