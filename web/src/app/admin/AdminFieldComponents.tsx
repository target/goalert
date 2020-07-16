import React, { useState } from 'react'
import Grid from '@material-ui/core/Grid'
import IconButton from '@material-ui/core/IconButton'
import Input from '@material-ui/core/Input'
import InputAdornment from '@material-ui/core/InputAdornment'
import Switch from '@material-ui/core/Switch'
import Visibility from '@material-ui/icons/Visibility'
import VisibilityOff from '@material-ui/icons/VisibilityOff'
import TelTextField from '../util/TelTextField'

interface InputProps {
  type: string
  name: string
  value: string
  password?: boolean
  onChange: (value: null | string) => void
  autoComplete?: string
}

export const StringListInput = (props: InputProps): JSX.Element => {
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

export function StringInput(props: InputProps): JSX.Element {
  const [showPassword, setShowPassword] = useState(false)
  const { onChange, password, type = 'text', ...rest } = props

  const renderPasswordAdornment = (): JSX.Element | null => {
    if (!props.password) return null

    return (
      <InputAdornment position='end'>
        <IconButton
          aria-label='Toggle password visibility'
          onClick={() => setShowPassword(!showPassword)}
        >
          {showPassword ? <Visibility /> : <VisibilityOff />}
        </IconButton>
      </InputAdornment>
    )
  }

  function handleChange(number: string): void {
    onChange('+' + number)
  }

  if (type === 'tel') {
    return (
      <TelTextField fullWidth InputProps onChange={handleChange} {...rest} />
    )
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

export const IntegerInput = (props: InputProps): JSX.Element => (
  <Input
    name={props.name}
    value={props.value}
    autoComplete={props.autoComplete}
    type='number'
    fullWidth
    onChange={(e) => props.onChange(e.target.value)}
  />
)

export const BoolInput = (props: InputProps): JSX.Element => (
  <Switch
    name={props.name}
    value={props.value}
    checked={props.value === 'true'}
    onChange={(e) => props.onChange(e.target.checked ? 'true' : 'false')}
  />
)
