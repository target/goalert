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
import useCreateAlerts from './useCreateAlerts'

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

  const [
    createAlerts,
    { data: alertsCreated, error: alertsFailed, loading: isCreatingAlerts },
  ] = useCreateAlerts(formFields)

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

  const onClose = () => {
    props.handleRequestClose()
    // NOTE dialog takes time to fade out
    setTimeout(() => {
      setActiveStep(0)
      setFormFields({
        summary: '',
        details: '',
        selectedServices: [],
        searchQuery: '',
      })
    }, 1000)
  }

  return (
    <Dialog
      open={props.open}
      onClose={onClose}
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
            mutationStatus={{ alertsCreated, alertsFailed, isCreatingAlerts }}
            onChange={e => onStepContentChange(e)}
          />
        </FormContainer>
      </DialogContent>
      <DialogNavigation
        activeStep={activeStep}
        formFields={formFields}
        onClose={onClose}
        onLastStep={onLastStep}
        onSubmit={createAlerts}
        setActiveStep={setActiveStep}
        steps={steps}
      />
    </Dialog>
  )
}
