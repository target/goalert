import React, { useState, MouseEvent } from 'react'
import Button from '@mui/material/Button'
import IconButton from '@mui/material/IconButton'
import Popover from '@mui/material/Popover'
import FilterList from '@mui/icons-material/FilterList'
import Box from '@mui/material/Box'
import SwipeableDrawer from '@mui/material/SwipeableDrawer'
import Switch from '@mui/material/Switch'
import Grid from '@mui/material/Grid'
import { useTheme } from '@mui/material/styles'
import { styles as globalStyles } from '../../styles/materialStyles'
import Radio from '@mui/material/Radio'
import RadioGroup from '@mui/material/RadioGroup'
import FormControlLabel from '@mui/material/FormControlLabel'
import FormControl from '@mui/material/FormControl'
import { useURLParam, useResetURLParams } from '../../actions'
import { useIsWidthDown } from '../../util/useWidth'

interface AlertsListFilterProps {
  serviceID: string
}

function AlertsListFilter(props: AlertsListFilterProps): React.JSX.Element {
  const theme = useTheme()
  const gs = globalStyles(theme)
  const [show, setShow] = useState(false)
  const [anchorEl, setAnchorEl] = useState<HTMLButtonElement | null>(null)

  const [filter, setFilter] = useURLParam<string>('filter', 'active')
  const [allServices, setAllServices] = useURLParam<boolean>(
    'allServices',
    false,
  )
  const [showAsFullTime, setShowAsFullTime] = useURLParam<boolean>(
    'fullTime',
    false,
  )
  const resetAll = useResetURLParams('filter', 'allServices', 'fullTime') // don't reset search param
  const isMobile = useIsWidthDown('md')

  function handleOpenFilters(event: MouseEvent<HTMLButtonElement>): void {
    setAnchorEl(event.currentTarget)
    setShow(true)
  }

  function handleCloseFilters(): void {
    setShow(false)
  }

  function renderFilters(): React.JSX.Element {
    let favoritesFilter = null
    if (!props.serviceID) {
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
      <Grid
        container
        spacing={2}
        sx={[
          { margin: '0.5em' },
          isMobile
            ? { width: 'fit-content' }
            : { width: '17em' },
        ]}
      >
        <Grid
          item
          xs={12}
          sx={{ display: 'flex', alignItems: 'center' }}
        >
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
            {isMobile && (
              <RadioGroup
                aria-label='Alert Status Filters'
                name='status-filters'
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
              >
                <FormControlLabel
                  value='active'
                  control={<Radio />}
                  label='Active'
                />
                <FormControlLabel
                  value='unacknowledged'
                  control={<Radio />}
                  label='Unacknowledged'
                />
                <FormControlLabel
                  value='acknowledged'
                  control={<Radio />}
                  label='Acknowledged'
                />
                <FormControlLabel
                  value='closed'
                  control={<Radio />}
                  label='Closed'
                />
                <FormControlLabel value='all' control={<Radio />} label='All' />
              </RadioGroup>
            )}
          </FormControl>
        </Grid>
        <Grid item xs={12} sx={gs.filterActions}>
          <Button onClick={resetAll}>Reset</Button>
          <Button onClick={handleCloseFilters}>Done</Button>
        </Grid>
      </Grid>
    )

    // renders a popover on desktop, and a swipeable drawer on mobile devices
    return (
      <React.Fragment>
        <Box sx={{ display: { xs: 'none', md: 'block' } }}>
          <Popover
            anchorEl={anchorEl}
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
        </Box>
        <Box sx={{ display: { xs: 'block', md: 'none' } }}>
          <SwipeableDrawer
            anchor='top'
            disableDiscovery
            disableSwipeToOpen
            open={show}
            onClose={handleCloseFilters}
            onOpen={handleOpenFilters}
            SlideProps={{
              unmountOnExit: true,
            }}
          >
            {content}
          </SwipeableDrawer>
        </Box>
      </React.Fragment>
    )
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
        onClick={handleOpenFilters}
        size='large'
      >
        <FilterList />
      </IconButton>
      {renderFilters()}
    </React.Fragment>
  )
}

export default AlertsListFilter
