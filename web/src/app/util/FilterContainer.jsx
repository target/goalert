import React, { useState } from 'react'
import p from 'prop-types'
import {
  Hidden,
  Popover,
  SwipeableDrawer,
  IconButton,
  Grid,
  Button,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { FilterList as FilterIcon } from '@mui/icons-material'

const useStyles = makeStyles((theme) => {
  return {
    actions: {
      display: 'flex',
      justifyContent: 'flex-end',
    },
    overflow: {
      overflow: 'visible',
    },
    container: {
      padding: 8,
      [theme.breakpoints.up('md')]: { width: '22em' },
      [theme.breakpoints.down('md')]: { width: '100%' },
    },
    formContainer: {
      margin: 0,
    },
  }
})

export default function FilterContainer(props) {
  const classes = useStyles()
  const [anchorEl, setAnchorEl] = useState(null)
  const {
    icon = <FilterIcon />,
    title = 'Filter',
    iconButtonProps,
    anchorRef,
  } = props

  function renderContent() {
    return (
      <Grid container spacing={2} className={classes.container}>
        <Grid
          item
          container
          xs={12}
          spacing={2}
          className={classes.formContainer}
        >
          {props.children}
        </Grid>
        <Grid item xs={12} className={classes.actions}>
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
      <Hidden mdDown>
        <Popover
          anchorEl={anchorEl}
          classes={{
            paper: classes.overflow,
          }}
          open={!!anchorEl}
          onClose={() => setAnchorEl(null)}
          TransitionProps={{
            unmountOnExit: true,
          }}
        >
          {renderContent()}
        </Popover>
      </Hidden>
      <Hidden mdUp>
        <SwipeableDrawer
          anchor='top'
          classes={{
            paper: classes.overflow,
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
      </Hidden>
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
