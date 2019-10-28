import React from 'react'
import { DialogActions, Button } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'

const mutation = gql`
  mutation CreateAlertMutation($input: CreateAlertInput!) {
    createAlert(input: $input) {
      id
    }
  }
`

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

  const [createAlert] = useMutation(mutation, {
    variables: {
      input: {
        serviceID: formFields.selectedServices[0],
        summary: formFields.summary.trim(),
        details: formFields.details.trim(),
      },
    },
  })

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
        onClick={onLastStep() ? () => createAlert() : handleNext}
        className={classes.button}
        disabled={nextIsDisabled(activeStep, formFields)}
      >
        {onLastStep() ? 'Submit' : 'Next'}
      </Button>
    </DialogActions>
  )
}
