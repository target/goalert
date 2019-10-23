import React, { useState } from 'react'
import {
  Dialog,
  Stepper,
  Step,
  StepLabel,
  DialogContent,
} from '@material-ui/core'
import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'
import { makeStyles } from '@material-ui/core/styles'
import classnames from 'classnames'
import DialogNavigation from './DialogNavigation'
import StepContent from './StepContent'
// import fetchServices from fetchServices

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

const useStyles = makeStyles(theme => ({
  instructions: {
    marginTop: theme.spacing(1),
    marginBottom: theme.spacing(1),
  },
}))

export default props => {
  const classes = useStyles()
  const [activeStep, setActiveStep] = useState(0)
  const [formFields, setFormFields] = useState({
    summary: '',
    details: '',
    searchQuery: '',
    services: [],
    labelKey: '',
    labelValue: '',
  })

  const { data, loading, error } = useQuery(query, {
    variables: { input: { search: formFields.searchQuery } },
  })

  // TODO refactor this hack
  // WANT: if (searchQuery changed?) { fetchAndUpdateServicesState() }
  if (!loading && !error) {
    const newState = { services: data.services.nodes }
    if (data.services.nodes !== formFields.services) {
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
      classes={{
        paper: classnames(classes.dialogWidth, classes.overflowVisible),
      }}
      c={console.log(formFields)}
    >
      <DialogContent className={classes.overflowVisible}>
        <Stepper activeStep={activeStep}>
          {steps.map(label => {
            return (
              <Step key={label}>
                <StepLabel>{label}</StepLabel>
              </Step>
            )
          })}
        </Stepper>
        <StepContent
          activeStep={activeStep}
          onChange={e => onStepContentChange(e)}
          formFields={formFields}
          setFormFields={setFormFields}
        />
        <DialogNavigation
          activeStep={activeStep}
          setActiveStep={setActiveStep}
          formFields={formFields}
          steps={steps}
        />
      </DialogContent>
    </Dialog>
  )
}
