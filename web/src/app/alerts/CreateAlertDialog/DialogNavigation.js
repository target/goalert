import React from 'react'
import { DialogActions, Button } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles(theme => ({
  button: {
    marginRight: theme.spacing(1),
  },
  dialogActions: {
    display: 'flex',
    flexDirection: 'row-reverse',
    justifyContent: 'flex-start',
  },
}))

export default function DialogNavigation(props) {
  const { activeStep, formFields, onClose, onSubmit, setActiveStep } = props
  const classes = useStyles()

  const stepForward = () => {
    setActiveStep(prevActiveStep => prevActiveStep + 1)
  }

  const stepBackward = () => {
    setActiveStep(prevActiveStep => prevActiveStep - 1)
  }

  // NOTE buttons are mounted in order of tab precedence and arranged with CSS
  // https://www.maxability.co.in/2016/06/13/tabindex-for-accessibility-good-bad-and-ugly/
  const renderButtons = () => {
    switch (activeStep) {
      case 0:
        return (
          <React.Fragment>
            <Button
              variant='contained'
              color='primary'
              onClick={stepForward}
              className={classes.button}
              disabled={!(formFields.summary && formFields.details)}
              type='button'
            >
              Next
            </Button>
            <Button onClick={onClose} className={classes.button}>
              Cancel
            </Button>
          </React.Fragment>
        )
      case 1:
        return (
          <React.Fragment>
            <Button
              variant='contained'
              color='primary'
              onClick={stepForward}
              className={classes.button}
              disabled={formFields.selectedServices.length === 0}
              type='button'
            >
              Next
            </Button>
            <Button onClick={stepBackward} className={classes.button}>
              Back
            </Button>
          </React.Fragment>
        )
      case 2:
        return (
          <React.Fragment>
            <Button
              variant='contained'
              color='primary'
              onClick={() => {
                onSubmit()
                stepForward()
              }}
              className={classes.button}
              disabled={formFields.selectedServices.length === 0}
              type='submit'
            >
              Next
            </Button>
            <Button onClick={stepBackward} className={classes.button}>
              Back
            </Button>
          </React.Fragment>
        )
      case 3:
        return (
          <React.Fragment>
            <Button
              variant='contained'
              color='primary'
              onClick={onClose}
              className={classes.button}
              type='button'
            >
              Done
            </Button>
          </React.Fragment>
        )
      default:
        return null
    }
  }

  return (
    <DialogActions className={classes.dialogActions}>
      {renderButtons()}
    </DialogActions>
  )
}
