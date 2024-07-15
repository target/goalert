import React, { useState } from 'react'
import {
  FormControl,
  FormControlLabel,
  FormLabel,
  Grid,
  Radio,
  RadioGroup,
  Stepper,
  Step,
  StepButton,
  TextField,
} from '@mui/material'
import { ActionInput, KeyRuleInput } from '../../../schema'
import UniversalKeyActionsList from './UniversalKeyActionsList'
import UniversalKeyActionsForm from './UniversalKeyActionsForm'

interface UniversalKeyRuleFormProps {
  value: KeyRuleInput
  onChange: (val: Readonly<KeyRuleInput>) => void

  nameError?: string
  descriptionError?: string
  conditionError?: string

  step: number
  setStep: (step: number) => void
}

const STEPS = ['Configure Rule', 'Configure Action(s)', 'Wrap Up']

export default function UniversalKeyRuleForm(
  props: UniversalKeyRuleFormProps,
): JSX.Element {
  const { step, setStep } = props
  const [editAction, setEditAction] = useState('')

  const handleChipClick = (action: ActionInput): void => {
    setEditAction(action.dest.type)
  }

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Stepper activeStep={step} alternativeLabel nonLinear>
          {STEPS.map((label, index) => (
            <Step key={label + index}>
              <StepButton color='inherit' onClick={() => setStep(index)}>
                {label}
              </StepButton>
            </Step>
          ))}
        </Stepper>
      </Grid>
      {step === 0 && (
        <React.Fragment>
          <Grid item xs={12}>
            <TextField
              fullWidth
              label='Name'
              name='name'
              value={props.value.name}
              onChange={(e) => {
                props.onChange({ ...props.value, name: e.target.value })
              }}
              error={!!props.nameError}
              helperText={props.nameError}
            />
          </Grid>
          <Grid item xs={12}>
            <TextField
              fullWidth
              label='Description'
              name='description'
              value={props.value.description}
              onChange={(e) => {
                props.onChange({
                  ...props.value,
                  description: e.target.value,
                })
              }}
              error={!!props.descriptionError}
              helperText={props.descriptionError}
            />
          </Grid>
          <Grid item xs={12}>
            <TextField
              fullWidth
              label='Condition (Expr syntax)'
              name='conditionExpr'
              multiline
              rows={3}
              value={props.value.conditionExpr}
              onChange={(e) => {
                props.onChange({
                  ...props.value,
                  conditionExpr: e.target.value,
                })
              }}
              error={!!props.conditionError}
              helperText={props.conditionError}
            />
          </Grid>
        </React.Fragment>
      )}

      {step === 1 && (
        <React.Fragment>
          <Grid item xs={12}>
            <UniversalKeyActionsList
              actions={props.value.actions}
              onDelete={(a) =>
                props.onChange({
                  ...props.value,
                  actions: props.value.actions.filter((v) => v !== a),
                })
              }
              onChipClick={handleChipClick}
            />
          </Grid>
          <Grid item xs={12}>
            <UniversalKeyActionsForm
              value={props.value.actions}
              onChange={(actions) =>
                props.onChange({ ...props.value, actions })
              }
              editActionId={editAction}
              onChipClick={handleChipClick}
            />
          </Grid>
        </React.Fragment>
      )}

      {step === 2 && (
        <React.Fragment>
          <Grid item xs={12}>
            <UniversalKeyActionsList actions={props.value.actions} noEdit />
          </Grid>
          <Grid item xs={12}>
            <FormControl>
              <FormLabel id='demo-row-radio-buttons-group-label'>
                After actions complete:
              </FormLabel>
              <RadioGroup
                row
                name='stop-or-continue'
                value={props.value.continueAfterMatch ? 'continue' : 'stop'}
              >
                <FormControlLabel
                  value='continue'
                  onChange={() =>
                    props.onChange({ ...props.value, continueAfterMatch: true })
                  }
                  control={<Radio />}
                  label='Continue processing rules'
                />
                <FormControlLabel
                  value='stop'
                  onChange={() =>
                    props.onChange({
                      ...props.value,
                      continueAfterMatch: false,
                    })
                  }
                  control={<Radio />}
                  label='Stop at this rule'
                />
              </RadioGroup>
            </FormControl>
          </Grid>
        </React.Fragment>
      )}
    </Grid>
  )
}
