import React from 'react'
import {
  Hidden,
  Popover,
  SwipeableDrawer,
  withStyles,
  IconButton,
  Grid,
  Button,
} from '@material-ui/core'
import { FilterList } from '@material-ui/icons'

const style = theme => {
  return {
    overflow: {
      overflow: 'visible',
    },

    container: {
      padding: 8,
      [theme.breakpoints.up('md')]: { width: '17em' },
      [theme.breakpoints.down('sm')]: { width: '100%' },
    },
    formContainer: {
      margin: 0,
    },
  }
}

@withStyles(style)
export default class FilterContainer extends React.PureComponent {
  state = {
    anchorEl: null,
  }

  renderContent() {
    return (
      <Grid item container spacing={3} className={this.props.classes.container}>
        <Grid
          item
          container
          xs={12}
          spacing={2}
          className={this.props.classes.formContainer}
        >
          {this.props.children}
        </Grid>
        <Grid item container justify='flex-end' xs={12}>
          {this.props.onReset && (
            <Button onClick={this.props.onReset}>Reset</Button>
          )}
          <Button
            onClick={() =>
              this.setState({
                anchorEl: null,
              })
            }
          >
            Done
          </Button>
        </Grid>
      </Grid>
    )
  }
  render() {
    const { classes } = this.props
    return (
      <React.Fragment>
        <IconButton
          color='inherit'
          onClick={e =>
            this.setState({
              anchorEl: e.target,
            })
          }
          title='filter'
          aria-expanded={Boolean(this.state.anchorEl)}
        >
          <FilterList />
        </IconButton>
        <Hidden smDown>
          <Popover
            anchorEl={this.state.anchorEl}
            classes={{
              paper: classes.overflow,
            }}
            open={!!this.state.anchorEl}
            onClose={() =>
              this.setState({
                anchorEl: null,
              })
            }
            anchorOrigin={{
              vertical: 'bottom',
              horizontal: 'right',
            }}
            transformOrigin={{
              vertical: 'top',
              horizontal: 'right',
            }}
          >
            {this.renderContent()}
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
            open={!!this.state.anchorEl}
            onClose={() =>
              this.setState({
                anchorEl: null,
              })
            }
            onOpen={() => {}}
          >
            {this.renderContent()}
          </SwipeableDrawer>
        </Hidden>
      </React.Fragment>
    )
  }
}
