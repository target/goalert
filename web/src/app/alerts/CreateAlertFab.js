import React, { useState } from 'react'
import classnames from 'classnames'
import { makeStyles } from '@material-ui/core'
import CreateAlertDialog from './CreateAlertDialog/CreateAlertDialog'
import Fab from '@material-ui/core/Fab'
import AddIcon from '@material-ui/icons/Add'
import Tooltip from '@material-ui/core/Tooltip'
import p from 'prop-types'

const useStyles = makeStyles(theme => ({
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
}))

export default function CreateAlertFab(props) {
  const classes = useStyles()
  const [open, setOpen] = useState(false)

  let fabOpen = classes.transitionUp
  // use set padding for the larger vertical height on warning snackbar
  if (props.showFavoritesWarning) {
    fabOpen = classes.warningTransitionUp
  }

  const transitionClass = props.transition ? fabOpen : classes.fabClose

  return (
    <React.Fragment>
      <Tooltip title='Create Alert' aria-label='Create Alert' placement='left'>
        <Fab
          data-cy='page-fab'
          className={classnames(classes.fab, transitionClass)}
          color='primary'
          onClick={() => setOpen(true)}
        >
          <AddIcon />
        </Fab>
      </Tooltip>
      {open && (
        <CreateAlertDialog
          onClose={() => setOpen(false)}
          serviceID={props.serviceID}
        />
      )}
    </React.Fragment>
  )
}

CreateAlertFab.propTypes = {
  serviceID: p.string,
  showFavoritesWarning: p.bool, // sets a larger vertical padding for the snackbar toast message
  transition: p.bool, // bool to transition fab up or down from snackbar notification
}
