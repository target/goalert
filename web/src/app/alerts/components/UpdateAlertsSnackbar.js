import React from 'react'
import { PropTypes as p } from 'prop-types'
import Snackbar from '@mui/material/Snackbar'
import SnackbarContent from '@mui/material/SnackbarContent'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import CloseIcon from '@mui/icons-material/Close'
import ErrorIcon from '@mui/icons-material/Error'
import IconButton from '@mui/material/IconButton'
import makeStyles from '@mui/styles/makeStyles'

const icon = {
  fontSize: 20,
}

const useStyles = makeStyles((theme) => ({
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
}))

function UpdateAlertsSnackbar({
  errorMessage,
  updateMessage,
  onClose,
  onExited,
  open,
}) {
  const classes = useStyles()

  function getMessage() {
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

  return (
    <Snackbar
      autoHideDuration={!errorMessage ? 3000 : null}
      open={open}
      onClose={onClose}
      TransitionProps={{
        onExited,
      }}
    >
      <SnackbarContent
        className={errorMessage ? classes.error : classes.success}
        message={getMessage()}
        action={[
          <IconButton
            key='close'
            aria-label='Close'
            color='inherit'
            onClick={onClose}
            size='large'
          >
            <CloseIcon className={classes.closeIcon} />
          </IconButton>,
        ]}
      />
    </Snackbar>
  )
}

UpdateAlertsSnackbar.propTypes = {
  errorMessage: p.string,
  open: p.bool.isRequired,
  updateMessage: p.string,
}

export default UpdateAlertsSnackbar
