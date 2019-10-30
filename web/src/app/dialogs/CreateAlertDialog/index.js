import React, { useState } from 'react'
import {
  Dialog,
  Stepper,
  Step,
  StepLabel,
  DialogContent,
} from '@material-ui/core'
import { isWidthDown } from '@material-ui/core/withWidth'
import DialogNavigation from './DialogNavigation'
import StepContent from './StepContent'
import { FormContainer } from '../../forms'
import DialogTitleWrapper from '../components/DialogTitleWrapper'
import useWidth from '../../util/useWidth'

export default props => {
  const width = useWidth()

  const [activeStep, setActiveStep] = useState(0)
  const [formFields, setFormFields] = useState({
    // data for mutation
    summary: '',
    details: '',
    selectedServices: [],

    // form helper
    searchQuery: '',
  })

  const onStepContentChange = e => {
    setFormFields(prevState => ({ ...prevState, ...e }))
  }

  const steps = [
    'Summary and Details',
    'Label Selection',
    'Confirm and Submit',
    null,
  ]

  const onLastStep = () => activeStep === steps.length - 1

  return (
    <Dialog
      open={props.open}
      onClose={props.handleRequestClose}
      fullScreen={isWidthDown('md', width)}
      fullWidth
      width={'md'}
    >
      <DialogTitleWrapper
        fullScreen={isWidthDown('md', width)}
        title={onLastStep() ? 'Review Created Alerts' : 'Create New Alert'}
      />
      <DialogContent>
        {!onLastStep() && (
          <Stepper activeStep={activeStep}>
            {steps.map(
              label =>
                label && (
                  <Step key={label}>
                    <StepLabel>{label}</StepLabel>
                  </Step>
                ),
            )}
          </Stepper>
        )}
        <FormContainer
          onChange={e => onStepContentChange(e)}
          value={formFields}
        >
          <StepContent
            activeStep={activeStep}
            formFields={formFields}
            onChange={e => onStepContentChange(e)}
          />
        </FormContainer>
      </DialogContent>
      <DialogNavigation
        activeStep={activeStep}
        setActiveStep={setActiveStep}
        formFields={formFields}
        steps={steps}
        onLastStep={onLastStep}
        onClose={props.handleRequestClose}
      />
    </Dialog>
  )
}
