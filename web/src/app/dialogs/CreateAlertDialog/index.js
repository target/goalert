import React, { useState } from 'react'
import {
  Dialog,
  Stepper,
  Step,
  StepLabel,
  DialogContent,
} from '@material-ui/core'
import { isWidthDown } from '@material-ui/core/withWidth'
import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'
import DialogNavigation from './DialogNavigation'
import StepContent from './StepContent'
import { FormContainer } from '../../forms'
import DialogTitleWrapper from '../components/DialogTitleWrapper'
import useWidth from '../../util/useWidth'

const query = gql`
  query($input: ServiceSearchOptions) {
    services(input: $input) {
      nodes {
        id
        name
        isFavorite
      }
    }
  }
`

const handleSubmit = () => {
  console.log('SUBMIT')
}

export default props => {
  const width = useWidth()

  const [activeStep, setActiveStep] = useState(0)
  const [formFields, setFormFields] = useState({
    // mutation data
    summary: '',
    details: '',
    selectedServices: [],

    // form helpers
    searchQuery: '',
    services: [],
  })

  const { data } = useQuery(query, {
    variables: {
      input: { search: formFields.searchQuery, favoritesFirst: true },
    },
    skip: formFields.searchQuery.length === 0,
    onCompleted: () => {
      const newState = { services: data.services.nodes }
      setFormFields(prevState => ({ ...prevState, ...newState }))
    },
    pollInterval: 0,
  })

  const onStepContentChange = e => {
    setFormFields(prevState => ({ ...prevState, ...e }))
  }

  const steps = ['Summary and Details', 'Label Selection', 'Create Alert']

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
