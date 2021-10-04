import React, { useState } from 'react'
import p from 'prop-types'
import Button from '@mui/material/Button'
import IconButton from '@mui/material/IconButton'
import Popover from '@mui/material/Popover'
import FilterList from '@mui/icons-material/FilterList'
import Hidden from '@mui/material/Hidden'
import SwipeableDrawer from '@mui/material/SwipeableDrawer'
import Switch from '@mui/material/Switch'
import Grid from '@mui/material/Grid'
import makeStyles from '@mui/styles/makeStyles';
import { styles as globalStyles } from '../../styles/materialStyles'
import Radio from '@mui/material/Radio'
import RadioGroup from '@mui/material/RadioGroup'
import FormControlLabel from '@mui/material/FormControlLabel'
import FormControl from '@mui/material/FormControl'
import classnames from 'classnames'
import { useURLParam, useResetURLParams } from '../../actions'
import useWidth from '../../util/useWidth'

// FIXME checkout https://mui.com/components/use-media-query/#migrating-from-withwidth
const withWidth = () => (WrappedComponent) => (props) => <WrappedComponent {...props} width="xs" />;

const useStyles = makeStyles((theme) => ({
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
}))

function AlertsListFilter({ serviceID }) {
  const classes = useStyles()
  const width = useWidth()
  const [show, setShow] = useState(false)
  const [anchorEl, setAnchorEl] = useState(null)

  const [filter, setFilter] = useURLParam('filter', 'active')
  const [allServices, setAllServices] = useURLParam('allServices', false)
  const [showAsFullTime, setShowAsFullTime] = useURLParam('fullTime', false)
  const resetAll = useResetURLParams('filter', 'allServices', 'fullTime') // don't reset search param

  function handleOpenFilters(event) {
    setAnchorEl(event.currentTarget)
    setShow(true)
  }

  function handleCloseFilters() {
    setShow(false)
  }

  function renderFilters() {
    // grabs class for width depending on breakpoints (md or higher uses popover width)
    const widthClass = isWidthUp('md', width) ? classes.popover : classes.drawer
    const gridClasses = classnames(classes.grid, widthClass)

    let favoritesFilter = null
    if (!serviceID) {
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
                  name='toggle-full-time'
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
          <Button onClick={handleCloseFilters}>Done</Button>
        </Grid>
      </Grid>
    )

    // renders a popover on desktop, and a swipeable drawer on mobile devices
    return (
      <React.Fragment>
        <Hidden mdDown>
          <Popover
            anchorEl={() => anchorEl}
            open={!!anchorEl && show}
            onClose={handleCloseFilters}
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
            open={show}
            onClose={handleCloseFilters}
            onOpen={handleOpenFilters}
          >
            {content}
          </SwipeableDrawer>
        </Hidden>
      </React.Fragment>
    );
  }

  /*
   * Finds the parent toolbar DOM node and appends the options
   * element to that node (after all the toolbar's children
   * are done being rendered)
   */

  return (
    <React.Fragment>
      <IconButton
        aria-label='Filter Alerts'
        color='inherit'
        onClick={handleOpenFilters}
        size="large">
        <FilterList />
      </IconButton>
      {renderFilters()}
    </React.Fragment>
  );
}

AlertsListFilter.propTypes = {
  serviceID: p.string,
}

export default AlertsListFilter
