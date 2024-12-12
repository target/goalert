import React from 'react'
import {
  FormControl,
  FormControlLabel,
  FormLabel,
  Grid,
  Radio,
  RadioGroup,
  TextField,
} from '@mui/material'
import { KeyRuleInput } from '../../../schema'
import { HelperText } from '../../forms'
import { ExprField } from './ExprField'

interface UniversalKeyRuleFormProps {
  value: KeyRuleInput
  onChange: (val: Readonly<KeyRuleInput>) => void

  nameError?: string
  descriptionError?: string
  conditionError?: string
}

export default function UniversalKeyRuleForm(
  props: UniversalKeyRuleFormProps,
): JSX.Element {
  return (
    <Grid container spacing={2}>
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
        <ExprField
          name='conditionExpr'
          label='Condition'
          value={props.value.conditionExpr}
          onChange={(v) => props.onChange({ ...props.value, conditionExpr: v })}
          error={!!props.conditionError}
        />
        <HelperText error={props.conditionError} />
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
    </Grid>
  )
}
