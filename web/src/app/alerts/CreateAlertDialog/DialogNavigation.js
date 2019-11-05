import React from 'react'
import { DialogActions, Button } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles(theme => ({
  button: {
    marginRight: theme.spacing(1),
  },
}))

const nextIsDisabled = (activeStep, formFields) => {
  switch (activeStep) {
    case 0:
      return !(formFields.summary && formFields.details)
    case 1:
      return formFields.selectedServices.length === 0
    default:
      return false
  }
}

export default props => {
  const {
    activeStep,
    formFields,
    onClose,
    onLastStep,
    onSubmit,
    setActiveStep,
    steps,
  } = props
  const classes = useStyles()

  const stepForward = () => {
    setActiveStep(prevActiveStep => prevActiveStep + 1)
  }

  const stepBackward = () => {
    setActiveStep(prevActiveStep => prevActiveStep - 1)
  }

  const getNextBtnLabel = () => {
    switch (activeStep) {
      case steps.length - 1:
        return 'Done'
      case steps.length - 2:
        return 'Submit'
      default:
        return 'Next'
    }
  }

  const getBackBtnLabel = () => {
    switch (activeStep) {
      case 0:
        return 'Cancel'
      default:
        return 'Back'
    }
  }

  const handleNext = () => {
    switch (activeStep) {
      case steps.length - 1:
        onClose()
        break
      case steps.length - 2:
        onSubmit()
        stepForward()
        break
      default:
        stepForward()
    }
  }

  const handleBack = () => {
    switch (activeStep) {
      case 0:
        return onClose()
      default:
        return stepBackward()
    }
  }

  return (
    <DialogActions>
      {!onLastStep() && (
        <Button onClick={handleBack} className={classes.button} tabIndex={-1}>
          {getBackBtnLabel()}
        </Button>
      )}

      <Button
        variant='contained'
        color='primary'
        onClick={handleNext}
        className={classes.button}
        disabled={nextIsDisabled(activeStep, formFields)}
      >
        {getNextBtnLabel()}
      </Button>
    </DialogActions>
  )
}
