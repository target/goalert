import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Button from '@material-ui/core/Button'
import AddIcon from '@material-ui/icons/Add'
import classnames from 'classnames'

import AlertForm from './AlertForm'
import withStyles from '@material-ui/core/styles/withStyles'

const styles = theme => ({
  fab: {
    position: 'fixed',
    bottom: '1em',
    right: '1em',
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
export default class CreateAlertFab extends Component {
  static propTypes = {
    serviceID: p.string, // used for alert form if on a service details page
    transition: p.bool, // bool to transition fab up or down from snackbar notification
  }

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

    return [
      <Button
        key='create-button'
        data-cy='page-fab'
        className={classnames(classes.fab, transitionClass)}
        variant='fab'
        color='primary'
        onClick={() => this.handleShowForm(true)}
      >
        <AddIcon />
      </Button>,
      <AlertForm
        key='alert-form'
        open={this.state.showForm}
        handleRequestClose={() => this.handleShowForm(false)}
        serviceID={this.props.serviceID}
      />,
    ]
  }
}
