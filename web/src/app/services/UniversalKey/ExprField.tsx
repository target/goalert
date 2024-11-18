import { FormControl, InputLabel, OutlinedInput } from '@mui/material'
import React from 'react'
import ExprEditor from '../../editor/ExprEditor'

export type ExprFieldProps = {
  error?: boolean
  name: string
  label: string
  value: string
  onChange: (value: string) => void
  dense?: boolean
  disabled?: boolean
}

export function ExprField(props: ExprFieldProps): React.ReactNode {
  const [isFocused, setIsFocused] = React.useState(false)
  return (
    <FormControl
      fullWidth
      error={!!props.error}
      variant='outlined'
      data-testid={'code-' + props.name}
    >
      <InputLabel htmlFor={props.name} shrink>
        {props.label} (Expr syntax)
      </InputLabel>
      <OutlinedInput
        name={props.name}
        id={props.name}
        label={props.label + ' (Expr syntax)'} // used for sizing, not display
        notched
        sx={{ padding: '1em' }}
        disabled={props.disabled}
        slots={{
          input: ExprEditor,
        }}
        inputProps={{
          maxHeight: props.dense && !isFocused ? '2em' : '20em',
          minHeight: props.dense && !isFocused ? '2em' : '5em',
          value: props.value,
          onFocus: () => setIsFocused(true),
          onBlur: () => setIsFocused(false),
          readOnly: props.disabled,

          // fix incorrect type infrence
          onChange: (v) => props.onChange(v as unknown as string),
        }}
      />
    </FormControl>
  )
}
