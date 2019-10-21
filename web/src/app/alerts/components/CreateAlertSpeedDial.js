import React, { Component } from 'react'
import AddIcon from '@material-ui/icons/Add'
import classnames from 'classnames'
import AlertForm from './AlertForm'
import withStyles from '@material-ui/core/styles/withStyles'
import SpeedDial from '../../util/SpeedDial'

const styles = theme => ({
  fab: {
    position: 'fixed',
    bottom: '2em',
    right: '2em',
    zIndex: 9001,
  },
  transitionUp: {
    transform: 'translate3d(0, -62px, 0)',
    transition: theme.transitions.create('transform', {
      duration: theme.transitions.duration.enteringScreen,
      easing: theme.transitions.easing.easeOut,
    }),
  },
  warningTransitionUp: {
    transform: 'translate3d(0, -7.75em, 0)',
    transition: theme.transitions.create('transform', {
      duration: theme.transitions.duration.enteringScreen,
      easing: theme.transitions.easing.easeOut,
    }),
  },
  fabClose: {
    transform: 'translate3d(0, 0, 0)',
    transition: theme.transitions.create('transform', {
      duration: theme.transitions.duration.leavingScreen,
      easing: theme.transitions.easing.sharp,
    }),
  },
})

@withStyles(styles)
export default class CreateAlertSpeedDial extends Component {
  state = {
    showForm: false,
  }

  handleShowForm = bool => {
    this.setState({
      showForm: bool,
    })
  }

  render() {
    const { classes, showFavoritesWarning, transition } = this.props

    let fabOpen = classes.transitionUp
    // use set padding for the larger verticle height on warning snackbar
    if (showFavoritesWarning) {
      fabOpen = classes.warningTransitionUp
    }

    const transitionClass = transition ? fabOpen : classes.fabClose

    return (
      <React.Fragment>
        <SpeedDial
          data-cy='page-speed-dial'
          color='primary'
          className={classnames(classes.fab, transitionClass)}
          label='chill'
          actions={[
            {
              label: 'Create Alert',
              onClick: () => this.handleShowForm(true),
              icon: <AddIcon />,
            },
          ]}
        />
        <AlertForm
          open={this.state.showForm}
          handleRequestClose={() => this.handleShowForm(false)}
        />
      </React.Fragment>
    )
  }
}
