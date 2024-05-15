import React from 'react'
import { FormContainer, FormField } from '../../forms'
import { Grid, TextField } from '@mui/material'
import { Rule } from '@mui/icons-material'

type Rule = {
  name: string
  expr: string
}

interface UniversalKeyRuleFormProps {
  value: Rule
  onChange: (val: Rule) => void
}

export default function UniversalKeyRuleForm(
  props: UniversalKeyRuleFormProps,
): JSX.Element {
  console.info(props.value.expr)

  return (
    <FormContainer value={props.value} onChange={props.onChange}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='Name'
            name='Name'
            required
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='Expr'
            name='Expr'
            required
            multiline
            rows={3}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
