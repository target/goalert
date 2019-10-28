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
import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'

const mutation = gql`
  mutation CreateAlertMutation($input: CreateAlertInput!) {
    createAlert(input: $input) {
      id
    }
  }
`

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

  const handleSubmit = () => {
    console.log('SUBMIT')

    formFields.selectedServices.forEach((serviceId, i) => {
      useMutation(mutation, {
        variables: {
          input: {
            service_id: serviceId,
            summary: formFields.summary.trim(),
            details: formFields.details.trim(),
            // description:
            //   formFields.summary.trim() + '\n' + formFields.details.trim(),
          },
        },
        onCompleted: data => console.log(data),
      })
    })
  }

  const steps = ['Summary and Details', 'Label Selection', 'Confirm and Submit']

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
        title='Create New Alert'
      />
      <DialogContent>
        <Stepper activeStep={activeStep}>
          {steps.map(label => (
            <Step key={label}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>
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
        handleSubmit={handleSubmit}
      />
    </Dialog>
  )
}
