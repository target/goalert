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

  const handleNext = () => {
    setActiveStep(prevActiveStep => prevActiveStep + 1)
  }

  const handleBack = () => {
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

  const onNextBtnClick = () => {
    switch (activeStep) {
      case steps.length - 1:
        onClose()
        break
      case steps.length - 2:
        onSubmit()
        handleNext()
        break
      default:
        handleNext()
    }
  }

  return (
    <DialogActions>
      {!onLastStep() && (
        <Button
          disabled={activeStep === 0}
          onClick={handleBack}
          className={classes.button}
        >
          Back
        </Button>
      )}

      <Button
        variant='contained'
        color='primary'
        onClick={onNextBtnClick}
        className={classes.button}
        disabled={nextIsDisabled(activeStep, formFields)}
      >
        {getNextBtnLabel()}
      </Button>
    </DialogActions>
  )
}
