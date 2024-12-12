import { FormControl, InputLabel, OutlinedInput } from '@mui/material'
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined'
import Tooltip from '@mui/material/Tooltip'
import React from 'react'
import ExprEditor from '../../editor/ExprEditor'
import AppLink from '../../util/AppLink'

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
      <InputLabel htmlFor={props.name} sx={{ padding: '0 50px 0 0' }} shrink>
        <div style={{ display: 'flex', alignItems: 'center' }}>
          {props.label}
          <AppLink
            to='https://expr-lang.org/docs/language-definition'
            newTab
            style={{ marginTop: '5px', marginRight: '5px', marginLeft: '5px' }}
          >
            <Tooltip title='Expr syntax'>
              <InfoOutlinedIcon />
            </Tooltip>
          </AppLink>
        </div>
      </InputLabel>
      <OutlinedInput
        name={props.name}
        id={props.name}
        label={props.label + ' ((i)) '} // used for sizing, not display
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
