import { FormControl, InputLabel, OutlinedInput } from '@mui/material'
import React from 'react'
import ExprEditor from '../../editor/ExprEditor'

export type ExprFieldProps = {
  error?: string
  name: string
  label: string
  value: string
  onChange: (value: string) => void
}

export function ExprField(props: ExprFieldProps): React.ReactNode {
  return (
    <FormControl fullWidth error={!!props.error} variant='outlined'>
      <InputLabel htmlFor={props.name} shrink>
        {props.label} (Expr syntax)
      </InputLabel>
      <OutlinedInput
        name={props.name}
        id={props.name}
        label={props.label + ' (Expr syntax)'} // used for sizing, not display
        notched
        sx={{ padding: '1em' }}
        slots={{
          input: ExprEditor,
        }}
        inputProps={{
          maxHeight: '10em',
          minHeight: '5em',
          value: props.value,

          // fix incorrect type infrence
          onChange: (v) => props.onChange(v as unknown as string),
        }}
      />
    </FormControl>
  )
}
