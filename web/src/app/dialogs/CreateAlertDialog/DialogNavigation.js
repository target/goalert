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
      return true
  }
}

export default props => {
  const { activeStep, formFields, setActiveStep, steps, handleSubmit } = props
  const classes = useStyles()

  const handleNext = () => {
    setActiveStep(prevActiveStep => prevActiveStep + 1)
  }

  const handleBack = () => {
    setActiveStep(prevActiveStep => prevActiveStep - 1)
  }

  return (
    <DialogActions>
      <Button
        disabled={activeStep === 0}
        onClick={handleBack}
        className={classes.button}
      >
        Back
      </Button>

      <Button
        variant='contained'
        color='primary'
        onClick={handleNext}
        className={classes.button}
        disabled={nextIsDisabled(activeStep, formFields)}
      >
        Next
      </Button>

      <Button
        variant='contained'
        color='primary'
        onClick={handleSubmit}
        className={classes.button}
        disabled={activeStep !== steps.length - 1}
      >
        Submit
      </Button>
    </DialogActions>
  )
}
