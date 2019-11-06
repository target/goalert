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
import { FormContainer, Form } from '../../forms'
import DialogTitleWrapper from '../../dialogs/components/DialogTitleWrapper'
import useWidth from '../../util/useWidth'
import useCreateAlerts from './useCreateAlerts'
import { fieldErrors } from '../../util/errutil'

export default props => {
  const width = useWidth()

  const [activeStep, setActiveStep] = useState(0)
  const [formFields, setFormFields] = useState({
    // data for mutation
    Summary: '',
    Details: '',
    selectedServices: [],

    // form helper
    searchQuery: '',
  })

  const [
    createAlerts,
    { data: alertsCreated, error: alertsFailed, loading: isCreatingAlerts },
  ] = useCreateAlerts(formFields)

  const errors = fieldErrors(alertsFailed)

  if (activeStep !== 0 && errors.length > 0) {
    errors.forEach(err => {
      if (err.field === 'summary' || err.field === 'details') {
        setActiveStep(0)
      }
    })
  }

  const onStepContentChange = e => {
    setFormFields(prevState => ({ ...prevState, ...e }))
  }

  const steps = ['Alert Info', 'Service Selection', 'Review', null]

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
      open
      onClose={onLastStep() ? null : onClose} // NOTE only close on last step if user hits Done
      fullScreen={isWidthDown('md', width)}
      fullWidth
      width='md'
      PaperProps={
        isWidthDown('md', width)
          ? null
          : {
              style: {
                height: '65vh',
              },
            }
      }
    >
      <DialogTitleWrapper
        fullScreen={isWidthDown('md', width)}
        title={'Create New Alert'}
      />
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
      <DialogContent>
        <FormContainer
          onChange={e => onStepContentChange(e)}
          value={formFields}
          errors={errors}
          optionalLabels
        >
          <Form id='create-alert-form'>
            <StepContent
              activeStep={activeStep}
              formFields={formFields}
              mutationStatus={{ alertsCreated, alertsFailed, isCreatingAlerts }}
              onChange={e => onStepContentChange(e)}
            />
          </Form>
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
