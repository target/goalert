import React from 'react'
import { DialogActions, Button } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import useCreateAlerts from './useCreateAlerts'

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
  const { activeStep, formFields, setActiveStep, steps } = props
  const classes = useStyles()

  const handleNext = () => {
    setActiveStep(prevActiveStep => prevActiveStep + 1)
  }

  const handleBack = () => {
    setActiveStep(prevActiveStep => prevActiveStep - 1)
  }

  const onLastStep = () => activeStep === steps.length - 1

  const [createAlerts] = useCreateAlerts(formFields)

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
        onClick={onLastStep() ? () => createAlerts() : handleNext}
        className={classes.button}
        disabled={nextIsDisabled(activeStep, formFields)}
      >
        {onLastStep() ? 'Submit' : 'Next'}
      </Button>
    </DialogActions>
  )
}
