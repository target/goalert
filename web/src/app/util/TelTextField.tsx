import React, { useEffect, useState } from 'react'
import { useQuery, gql } from 'urql'
import TextField, { TextFieldProps } from '@mui/material/TextField'
import { InputProps } from '@mui/material/Input'
import { Check, Close } from '@mui/icons-material'
import _ from 'lodash'
import InputAdornment from '@mui/material/InputAdornment'
import makeStyles from '@mui/styles/makeStyles'
import { DEBOUNCE_DELAY } from '../config'

const isValidNumber = gql`
  query ($number: String!) {
    phoneNumberInfo(number: $number) {
      id
      valid
    }
  }
`

const useStyles = makeStyles({
  valid: {
    fill: 'green',
  },
  invalid: {
    fill: 'red',
  },
})

const noSuspense = { suspense: false }

export default function TelTextField(
  props: TextFieldProps & { value: string },
): JSX.Element {
  const classes = useStyles()
  const [phoneNumber, setPhoneNumber] = useState('')

  // debounce to set the phone number
  useEffect(() => {
    const t = setTimeout(() => {
      setPhoneNumber(props.value)
    }, DEBOUNCE_DELAY)
    return () => clearTimeout(t)
  }, [props.value])

  // check validation of the input phoneNumber through graphql
  const [{ data }] = useQuery({
    query: isValidNumber,
    variables: { number: phoneNumber },
    requestPolicy: 'cache-first',
    pause: !phoneNumber || props.disabled,
    context: noSuspense,
  })

  // fetch validation
  const valid = _.get(data, 'phoneNumberInfo.valid', null)

  let adorn
  if (!props.value) {
    // no adornment if empty
  } else if (valid) {
    adorn = <Check className={classes.valid} />
  } else if (valid === false) {
    adorn = <Close className={classes.invalid} />
  }

  let iprops: Partial<InputProps>
  iprops = {
    startAdornment: (
      <InputAdornment position='start' style={{ marginBottom: '0.1em' }}>
        +
      </InputAdornment>
    ),
  }

  // if has inputProps from parent component, spread it in the iprops
  if (props.InputProps !== undefined) {
    iprops = {
      ...iprops,
      ...props.InputProps,
    }
  }

  // add live validation icon to the right of the textfield as an endAdornment
  if (adorn) {
    iprops = {
      endAdornment: <InputAdornment position='end'>{adorn}</InputAdornment>,
      ...iprops,
    }
  }

  // remove unwanted character
  function handleChange(e: React.ChangeEvent<HTMLInputElement>): void {
    if (!props.onChange) return
    if (!e.target.value) return props.onChange(e)

    // ignore SID being pasted in
    if (e.target.value.toLowerCase().startsWith('mg')) return

    e.target.value = '+' + e.target.value.replace(/[^0-9]/g, '')
    return props.onChange(e)
  }

  return (
    <TextField
      fullWidth
      {...props}
      InputProps={iprops}
      type={props.type || 'tel'}
      helperText={
        props.helperText ||
        'Include country code e.g. +1 (USA), +91 (India), +44 (UK)'
      }
      onChange={handleChange}
      value={(props.value || '').replace(/[^0-9]/g, '')}
    />
  )
}
