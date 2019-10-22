import React, { Component } from 'react'
import AddIcon from '@material-ui/icons/Add'
import LabelIcon from '@material-ui/icons/Label'
import classnames from 'classnames'
import AlertForm from './AlertForm'
import CreateAlertByLabelDialog from './CreateAlertByLabelDialog'
import withStyles from '@material-ui/core/styles/withStyles'
import SpeedDial from '../../util/SpeedDial'

const styles = theme => ({
  speedDial: {
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
  speedDialClose: {
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
    showCreateAlertForm: false,
    showCreateAlertByLabelDialog: false,
  }

  showForm = form => {
    switch (form) {
      case 'createAlert':
        this.setState({
          showCreateAlertForm: true,
        })
        break
      case 'createAlertByLabel':
        this.setState({
          showCreateAlertByLabelDialog: true,
        })
        break
      default:
        this.setState({
          showCreateAlertForm: false,
          showCreateAlertByLabelDialog: false,
        })
        break
    }
  }

  render() {
    const { classes, showFavoritesWarning, transition } = this.props

    let speedDialOpen = classes.transitionUp
    // use set padding for the larger verticle height on warning snackbar
    if (showFavoritesWarning) {
      speedDialOpen = classes.warningTransitionUp
    }

    const transitionClass = transition ? speedDialOpen : classes.speedDialClose

    return (
      <React.Fragment>
        <SpeedDial
          data-cy='page-speed-dial'
          color='primary'
          className={classnames(classes.speedDial, transitionClass)}
          label='Create Alert'
          actions={[
            {
              label: 'Single Service',
              onClick: () => this.showForm('createAlert'),
              icon: <AddIcon />,
            },
            {
              label: 'Multi-Service',
              onClick: () => this.showForm('createAlertByLabel'),
              icon: <LabelIcon />,
            },
          ]}
        />
        <AlertForm
          open={this.state.showCreateAlertForm}
          handleRequestClose={() => this.showForm(null)}
        />
        <CreateAlertByLabelDialog
          open={this.state.showCreateAlertByLabelDialog}
          handleRequestClose={() => this.showForm(null)}
        />
      </React.Fragment>
    )
  }
}
