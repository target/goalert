import React, { useState } from 'react'
import {
  Grid,
  TextField,
  Dialog,
  Stepper,
  Step,
  StepLabel,
  DialogContent,
  InputAdornment,
} from '@material-ui/core'
// import gql from 'graphql-tag'

import { makeStyles } from '@material-ui/core/styles'
// import Typography from '@material-ui/core/Typography'
import classnames from 'classnames'
import { FormContainer, FormField } from '../../forms'
import ServiceLabelFilterContainer from '../../services/ServiceLabelFilterContainer'
import { Search as SearchIcon } from '@material-ui/icons'
import DialogNavigation from './DialogNavigation'

const useStyles = makeStyles(theme => ({
  instructions: {
    marginTop: theme.spacing(1),
    marginBottom: theme.spacing(1),
  },
}))

const StepContent = props => {
  switch (props.activeStep) {
    case 0:
      return (
        <Grid container spacing={2}>
          <FormContainer onChange={props.onChange}>
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
          </FormContainer>
        </Grid>
      )
    case 1:
      return (
        <Grid item xs={12}>
          <FormContainer onChange={props.onChange}>
            {/* <Grid item xs={12}>
              <SelectedServices />
            </Grid> */}
            <FormField
              fullWidth
              label='Search Query'
              name='searchQuery'
              fieldName='searchQuery'
              required
              component={TextField}
              InputProps={{
                // ref: fieldRef,
                startAdornment: (
                  <InputAdornment position='start'>
                    <SearchIcon color='action' />
                  </InputAdornment>
                ),
                endAdornment: (
                  <ServiceLabelFilterContainer
                    // anchorRef={fieldRef}
                    labelKey={'key'}
                    labelValue={'() => console.log(value)'}
                    onKeyChange={() => console.log('onKeyChange')}
                    onValueChange={() => console.log('onValueChange')}
                    // onReset={() => setSearchParam()}
                  />
                ),
              }}
            />
          </FormContainer>
        </Grid>
      )

    case 2:
      return 'plz confirm ur info'
    default:
      return 'Unknown step'
  }
}

export default function CreateAlertByLabelDialog(props) {
  const classes = useStyles()
  const [activeStep, setActiveStep] = useState(0)
  const [formFields, setFormFields] = useState({
    summary: '',
    details: '',
    searchQuery: '',
    serviceIds: [],
  })

  const onStepContentChange = e => {
    setFormFields(prevState => ({ ...prevState, ...e }))
  }

  const steps = ['Summary and Details', 'Label Selection', 'Create Alert']

  return (
    <Dialog
      open={props.open}
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
