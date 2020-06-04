import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Snackbar from '@material-ui/core/Snackbar'
import SnackbarContent from '@material-ui/core/SnackbarContent'
import CheckCircleIcon from '@material-ui/icons/CheckCircle'
import CloseIcon from '@material-ui/icons/Close'
import ErrorIcon from '@material-ui/icons/Error'
import IconButton from '@material-ui/core/IconButton'
import withStyles from '@material-ui/core/styles/withStyles'

const icon = {
  fontSize: 20,
}

const styles = (theme) => ({
  success: {
    backgroundColor: 'green',
  },
  error: {
    backgroundColor: theme.palette.error.dark,
  },
  closeIcon: {
    ...icon,
  },
  resultIcon: {
    ...icon,
    opacity: 0.9,
    marginRight: theme.spacing(1),
  },
  message: {
    display: 'flex',
    alignItems: 'center',
  },
})

@withStyles(styles)
export default class UpdateAlertsSnackbar extends Component {
  static propTypes = {
    errorMessage: p.string,
    open: p.bool.isRequired,
    updateMessage: p.string,
  }

  getMessage = () => {
    const { classes, errorMessage, updateMessage } = this.props

    if (errorMessage) {
      return (
        <span className={classes.message}>
          <ErrorIcon className={classes.resultIcon} />
          {errorMessage}
        </span>
      )
    }
    return (
      <span className={classes.message} data-cy='update-message'>
        <CheckCircleIcon className={classes.resultIcon} />
        {updateMessage}
      </span>
    )
  }

  render() {
    const { classes, errorMessage, onClose, onExited, open } = this.props

    return (
      <Snackbar
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
        autoHideDuration={!errorMessage ? 3000 : null}
        open={open}
        onClose={onClose}
        onExited={onExited}
      >
        <SnackbarContent
          className={errorMessage ? classes.error : classes.success}
          message={this.getMessage()}
          action={[
            <IconButton
              key='close'
              aria-label='Close'
              color='inherit'
              onClick={onClose}
            >
              <CloseIcon className={classes.closeIcon} />
            </IconButton>,
          ]}
        />
      </Snackbar>
    )
  }
}
