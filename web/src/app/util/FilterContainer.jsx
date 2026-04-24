import React, { useState } from 'react'
import p from 'prop-types'
import {
  Box,
  Popover,
  SwipeableDrawer,
  IconButton,
  Grid,
  Button,
} from '@mui/material'
import { FilterList as FilterIcon } from '@mui/icons-material'

export default function FilterContainer(props) {
  const [anchorEl, setAnchorEl] = useState(null)
  const {
    icon = <FilterIcon />,
    title = 'Filter',
    iconButtonProps,
    anchorRef,
  } = props

  function renderContent() {
    return (
      <Grid
        container
        spacing={2}
        sx={(theme) => ({
          padding: 1,
          [theme.breakpoints.up('md')]: { width: '22em' },
          [theme.breakpoints.down('md')]: { width: '100%' },
        })}
      >
        <Grid size={12} container spacing={2} sx={{ margin: 0 }}>
          {props.children}
        </Grid>
        <Grid size={12} sx={{ display: 'flex', justifyContent: 'flex-end' }}>
          {props.onReset && (
            <Button data-cy='filter-reset' onClick={props.onReset}>
              Reset
            </Button>
          )}
          <Button data-cy='filter-done' onClick={() => setAnchorEl(null)}>
            Done
          </Button>
        </Grid>
      </Grid>
    )
  }

  return (
    <React.Fragment>
      <IconButton
        onClick={(e) => setAnchorEl(anchorRef ? anchorRef.current : e.target)}
        title={title}
        aria-expanded={Boolean(anchorEl)}
        {...iconButtonProps}
        size='large'
      >
        {icon}
      </IconButton>
      <Box sx={{ display: { xs: 'none', md: 'block' } }}>
        <Popover
          anchorEl={anchorEl}
          slotProps={{
            paper: { sx: { overflow: 'visible' } },
          }}
          open={!!anchorEl}
          onClose={() => setAnchorEl(null)}
          TransitionProps={{
            unmountOnExit: true,
          }}
        >
          {renderContent()}
        </Popover>
      </Box>
      <Box sx={{ display: { xs: 'block', md: 'none' } }}>
        <SwipeableDrawer
          anchor='top'
          slotProps={{
            paper: { sx: { overflow: 'visible' } },
          }}
          disableDiscovery
          disableSwipeToOpen
          open={!!anchorEl}
          onClose={() => setAnchorEl(null)}
          onOpen={() => {}}
          SlideProps={{
            unmountOnExit: true,
          }}
        >
          {renderContent()}
        </SwipeableDrawer>
      </Box>
    </React.Fragment>
  )
}

FilterContainer.propTypes = {
  icon: p.node,
  // https://material-ui.com/api/icon-button/
  iconButtonProps: p.object,
  onReset: p.func,
  title: p.string,

  anchorRef: p.object,
  children: p.node,
}
