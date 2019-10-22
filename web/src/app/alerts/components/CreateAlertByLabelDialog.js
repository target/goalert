import React, { useState } from 'react'
import {
  Grid,
  TextField,
  Dialog,
  DialogActions,
  Stepper,
  Step,
  StepLabel,
  Button,
} from '@material-ui/core'
// import gql from 'graphql-tag'

import { makeStyles } from '@material-ui/core/styles'
// import Typography from '@material-ui/core/Typography'
import classnames from 'classnames'
import { FormContainer, FormField } from '../../forms'

const useStyles = makeStyles(theme => ({
  button: {
    marginRight: theme.spacing(1),
  },
  instructions: {
    marginTop: theme.spacing(1),
    marginBottom: theme.spacing(1),
  },
}))

const StepContent = props => {
  switch (props.activeStep) {
    case 0:
      return (
        <FormContainer onChange={props.onChange}>
          <Grid container spacing={2}>
            <Grid item xs={12}>
              <FormField
                fullWidth
                label='Alert Summary'
                name='summary'
                required
                component={TextField}
              />
            </Grid>
            <Grid item xs={12}>
              <FormField
                fullWidth
                label='Alert Details'
                name='details'
                required
                component={TextField}
              />
            </Grid>
          </Grid>
        </FormContainer>
      )
    case 1:
      return (
        <FormContainer onChange={props.onChange}>
          <Grid container spacing={2}>
            <Grid item xs={12}>
              {/* <Selection Chips w/ delete btns /> */}
            </Grid>
          </Grid>
        </FormContainer>
      )
    case 2:
      return 'This is the bit I really care about!'
    default:
      return 'Unknown step'
  }
}

const nextIsDisabled = (activeStep, formFields) => {
  switch (activeStep) {
    case 0:
      return !(formFields.summary && formFields.details)
    case 1:
      return !formFields.serviceIds.length
    // case 2:
    //   return !
    default:
      return false
  }
}

export default function CreateAlertByLabelDialog(props) {
  const classes = useStyles()
  const [activeStep, setActiveStep] = useState(0)
  const [formFields, setFormFields] = useState({
    summary: '',
    details: '',
    serviceIds: [],
  })

  const onStepContentChange = e => {
    setFormFields(prevState => ({ ...prevState, ...e }))
  }

  const steps = ['Summary and Details', 'Label Selection', 'Create Alert']

  const handleNext = () => {
    setActiveStep(prevActiveStep => prevActiveStep + 1)
  }

  const handleBack = () => {
    setActiveStep(prevActiveStep => prevActiveStep - 1)
  }

  const handleSubmit = () => {
    console.log('SUBMIT')
  }

  return (
    <Dialog
      open={props.open}
      onClose={props.handleRequestClose}
      classes={{
        paper: classnames(classes.dialogWidth, classes.overflowVisible),
      }}
      c={console.log(formFields)}
    >
      <Stepper activeStep={activeStep}>
        {steps.map((label, index) => {
          const stepProps = {}
          const labelProps = {}
          return (
            <Step key={label} {...stepProps}>
              <StepLabel {...labelProps}>{label}</StepLabel>
            </Step>
          )
        })}
      </Stepper>
      <div>
        <StepContent
          activeStep={activeStep}
          onChange={e => onStepContentChange(e)}
        />
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
      </div>
    </Dialog>
  )
}
