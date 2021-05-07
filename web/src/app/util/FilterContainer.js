import React, { useState } from 'react'
import p from 'prop-types'
import {
  Hidden,
  Popover,
  SwipeableDrawer,
  IconButton,
  Grid,
  Button,
  makeStyles,
} from '@material-ui/core'
import { FilterList as FilterIcon } from '@material-ui/icons'

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
      [theme.breakpoints.down('sm')]: { width: '100%' },
    },
    formContainer: {
      margin: 0,
    },
  }
})

export default function FilterContainer(props) {
  const classes = useStyles()
  const [anchorEl, setAnchorEl] = useState(null)

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

  const { icon, iconButtonProps, anchorRef } = props
  return (
    <React.Fragment>
      <IconButton
        color='inherit'
        onClick={(e) => setAnchorEl(anchorRef ? anchorRef.current : e.target)}
        title={props.title}
        aria-expanded={Boolean(anchorEl)}
        {...iconButtonProps}
      >
        {icon}
      </IconButton>
      <Hidden smDown>
        <Popover
          anchorEl={anchorEl}
          classes={{
            paper: classes.overflow,
          }}
          open={!!anchorEl}
          onClose={() => setAnchorEl(null)}
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

FilterContainer.defaultProps = {
  icon: <FilterIcon />,
  title: 'Filter',
}
