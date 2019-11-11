import React, { useState } from 'react'
import {
  Dialog,
  Stepper,
  Step,
  StepLabel,
  DialogContent,
  makeStyles,
} from '@material-ui/core'
import { isWidthUp } from '@material-ui/core/withWidth'
import DialogNavigation from './DialogNavigation'
import StepContent from './StepContent'
import { FormContainer, Form } from '../../forms'
import DialogTitleWrapper from '../../dialogs/components/DialogTitleWrapper'
import useWidth from '../../util/useWidth'
import useCreateAlerts from './useCreateAlerts'
import { fieldErrors } from '../../util/errutil'

const useStyles = makeStyles(theme => ({
  dialog: {
    [theme.breakpoints.up('md')]: {
      height: '65vh',
    },
  },
}))

export default function CreateAlertDialog(props) {
  const width = useWidth()
  const classes = useStyles()

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

  const stepsTitles = ['Alert Info', 'Service Selection', 'Review', null]

  const onLastStep = activeStep === stepsTitles.length - 1

  const onClose = () => {
    props.handleRequestClose()
  }

  const resetForm = () => {
    setActiveStep(0)
    setFormFields({
      summary: '',
      details: '',
      selectedServices: [],
      searchQuery: '',
    })
  }

  const isWideScreen = isWidthUp('md', width)

  return (
    <Dialog
      open={props.open}
      onClose={onLastStep ? null : onClose} // NOTE only close on last step if user hits Done
      fullScreen={!isWideScreen}
      fullWidth
      width='md'
      PaperProps={{ className: classes.dialog }}
      onExited={resetForm}
    >
      <DialogTitleWrapper fullScreen={!isWideScreen} title='Create New Alert' />
      {!onLastStep && (
        <Stepper activeStep={activeStep}>
          {stepsTitles.map(
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
          errors={fieldErrors(alertsFailed)}
          optionalLabels
        >
          <Form id='create-alert-form'>
            <StepContent
              activeStep={activeStep}
              setActiveStep={setActiveStep}
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
        onSubmit={createAlerts}
        setActiveStep={setActiveStep}
      />
    </Dialog>
  )
}
