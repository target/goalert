import React from 'react'
import { FormContainer, FormField } from '../../forms'
import { Grid, TextField } from '@mui/material'
import { KeyRule } from '../../../schema'

interface UniversalKeyRuleFormProps {
  value: KeyRule
  onChange: (val: KeyRule) => void
}

export default function UniversalKeyRuleForm(
  props: UniversalKeyRuleFormProps,
): JSX.Element {
  return (
    <FormContainer value={props.value} onChange={props.onChange}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
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
            label='Description'
            name='description'
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='Expr'
            name='conditionExpr'
            required
            multiline
            rows={3}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
