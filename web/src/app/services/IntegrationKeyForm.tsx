import React, { useState } from 'react'
import Grid from '@mui/material/Grid'
import Stepper from '@mui/material/Stepper'
import Step from '@mui/material/Step'
import StepButton from '@mui/material/StepButton'
import StepContent from '@mui/material/StepContent'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import MenuItem from '@mui/material/MenuItem'
import { FormContainer, FormField } from '../forms'
import { IntegrationKeyType } from '../../schema'
import { FieldError } from '../util/errutil'
import { useFeatures } from '../util/RequireConfig'
import MaterialSelect from '../selection/MaterialSelect'

export interface Value {
  name: string
  type: IntegrationKeyType
  filters: Filter[]
  summary: string
  details: string
  deduplication: string[]
}

export interface Filter {
  key: string
  value: string
}

interface IntegrationKeyFormProps {
  value: Value

  errors: FieldError[]

  onChange: (val: Value) => void

  // can be deleted when FormContainer.js is converted to ts
  disabled: boolean
}

export default function IntegrationKeyForm(
  props: IntegrationKeyFormProps,
): JSX.Element {
  const [step, setStep] = useState(0)
  const types = useFeatures().integrationKeyTypes

  function handleStepChange(stepChange: number): void {
    if (stepChange === step) {
      setStep(-1) // close
    } else {
      setStep(stepChange) // open
    }
  }

  const { ...formProps } = props
  return (
    <FormContainer {...formProps} optionalLabels>
      <Stepper activeStep={step} nonLinear orientation='vertical'>
        <Step>
          <StepButton
            aria-expanded={step === 0}
            onClick={() => handleStepChange(0)}
            tabIndex={-1}
          >
            <Typography>Info</Typography>
          </StepButton>
          <StepContent>
            <Grid container spacing={2}>
              <Grid item style={{ flexGrow: 1 }} xs={12}>
                <FormField
                  fullWidth
                  component={TextField}
                  label='Name'
                  name='name'
                  required
                />
              </Grid>
              <Grid item xs={12}>
                <FormField
                  fullWidth
                  component={TextField}
                  select
                  required
                  label='Type'
                  name='type'
                >
                  {types.map((t) => (
                    <MenuItem disabled={!t.enabled} key={t.id} value={t.id}>
                      {t.name}
                    </MenuItem>
                  ))}
                </FormField>
              </Grid>
            </Grid>
          </StepContent>
        </Step>
        <Step>
          <StepButton
            aria-expanded={step === 1}
            onClick={() => handleStepChange(1)}
            tabIndex={-1}
            optional='Optional'
          >
            <Typography>Filter</Typography>
          </StepButton>
        </Step>
        <Step>
          <StepButton
            aria-expanded={step === 2}
            onClick={() => handleStepChange(2)}
            tabIndex={-1}
          >
            <Typography>Action</Typography>
          </StepButton>
          <StepContent>
            <Grid container spacing={2}>
              <Grid item style={{ flexGrow: 1 }} xs={12}>
                <FormField
                  fullWidth
                  component={TextField}
                  label='Summary'
                  name='summary'
                  placeholder='{{ body.alert }}'
                  required
                />
              </Grid>
              <Grid item style={{ flexGrow: 1 }} xs={12}>
                <FormField
                  fullWidth
                  component={TextField}
                  label='Details'
                  name='details'
                  placeholder='{{ body.status.code }}'
                  required
                />
              </Grid>
              <Grid item style={{ flexGrow: 1 }} xs={12}>
                <FormField
                  fullWidth
                  component={MaterialSelect}
                  multiple
                  options={formProps.value.filters.map((f) => ({
                    label: f.key,
                    value: f.key,
                  }))}
                  label='Deduplication'
                  name='deduplication'
                />
              </Grid>
            </Grid>
          </StepContent>
        </Step>
      </Stepper>
    </FormContainer>
  )
}
