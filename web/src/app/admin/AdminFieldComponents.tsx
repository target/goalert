import React, { useState } from 'react'
import Grid from '@mui/material/Grid'
import IconButton from '@mui/material/IconButton'
import Input from '@mui/material/OutlinedInput'
import InputAdornment from '@mui/material/InputAdornment'
import Switch from '@mui/material/Switch'
import Visibility from '@mui/icons-material/Visibility'
import VisibilityOff from '@mui/icons-material/VisibilityOff'
import TelTextField from '../util/TelTextField'

interface InputProps {
  type?: string
  name: string
  value: string
  password?: boolean
  onChange: (value: null | string) => void
  autoComplete?: string
}

export function StringInput(props: InputProps): React.JSX.Element {
  const [showPassword, setShowPassword] = useState(false)
  const { onChange, password, type = 'text', ...rest } = props

  const renderPasswordAdornment = (): React.JSX.Element | null => {
    if (!props.password) return null

    return (
      <InputAdornment position='end'>
        <IconButton
          aria-label='Toggle password visibility'
          onClick={() => setShowPassword(!showPassword)}
          size='large'
        >
          {showPassword ? <Visibility /> : <VisibilityOff />}
        </IconButton>
      </InputAdornment>
    )
  }

  if (props.name === 'Twilio.FromNumber') {
    return <TelTextField onChange={(e) => onChange(e.target.value)} {...rest} />
  }
  return (
    <Input
      fullWidth
      autoComplete='new-password' // chrome keeps autofilling them, this stops it
      type={password && !showPassword ? 'password' : type}
      onChange={(e) => onChange(e.target.value)}
      endAdornment={renderPasswordAdornment()}
      {...rest}
    />
  )
}

export const StringListInput = (props: InputProps): React.JSX.Element => {
  const value = props.value ? props.value.split('\n').concat('') : ['']
  return (
    <Grid container spacing={1}>
      {value.map((val, idx) => (
        <Grid key={idx} item xs={12}>
          <StringInput
            type={props.type}
            value={val}
            name={val ? props.name + '-' + idx : props.name + '-new-item'}
            onChange={(newVal) =>
              props.onChange(
                value
                  .slice(0, idx)
                  .concat(newVal || '', ...value.slice(idx + 1))
                  .filter((v: string) => v)
                  .join('\n'),
              )
            }
            autoComplete='new-password'
            password={props.password}
          />
        </Grid>
      ))}
    </Grid>
  )
}

export const IntegerInput = (props: InputProps): React.JSX.Element => (
  <Input
    name={props.name}
    value={props.value}
    autoComplete={props.autoComplete}
    type='number'
    fullWidth
    onChange={(e) => props.onChange(e.target.value)}
    inputProps={{
      min: 0,
      max: 9000,
    }}
  />
)

export const BoolInput = (props: InputProps): React.JSX.Element => (
  <Switch
    name={props.name}
    value={props.value}
    checked={props.value === 'true'}
    onChange={(e) => props.onChange(e.target.checked ? 'true' : 'false')}
  />
)
