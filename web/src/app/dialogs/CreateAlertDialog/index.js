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
import { makeStyles } from '@material-ui/core/styles'
// import classnames from 'classnames'
import DialogNavigation from './DialogNavigation'
import StepContent from './StepContent'
import { styles as globalStyles } from '../../styles/materialStyles'
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

const useStyles = makeStyles(theme => {
  const { dialogWidth } = globalStyles(theme)
  return {
    dialogWidth,
  }
})

const handleSubmit = () => {
  console.log('SUBMIT')
}

export default props => {
  const classes = useStyles()
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
    labelKey: '',
    labelValue: '',
  })

  const { data, loading, error } = useQuery(query, {
    variables: {
      input: { search: formFields.searchQuery, favoritesFirst: true },
    },
  })

  // TODO: refactor this hack
  // WANT: if (searchQuery changed?) { fetchAndUpdateServicesState() }
  if (!loading && !error) {
    if (data.services.nodes !== formFields.services) {
      const newState = { services: data.services.nodes }
      setFormFields(prevState => ({ ...prevState, ...newState }))
    }
  }

  const onStepContentChange = e => {
    setFormFields(prevState => ({ ...prevState, ...e }))
  }

  const steps = ['Summary and Details', 'Label Selection', 'Create Alert']

  return (
    <Dialog
      open={props.open}
      // open={true}
      onClose={props.handleRequestClose}
      // classes={{
      //   paper: classes.dialogWidth,
      // }}
      className={classes.dialogWidth}
      c={console.log(formFields)}
      fullScreen={isWidthDown('md', width)}
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
