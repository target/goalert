import React, { useEffect, useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import TextField, { TextFieldProps } from '@material-ui/core/TextField'
import { InputProps } from '@material-ui/core/Input'
import { Check, Close } from '@material-ui/icons'
import InputAdornment from '@material-ui/core/InputAdornment'
import { makeStyles } from '@material-ui/core'
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

type InputType = 'tel' | 'sid'
type TelTextFieldProps = TextFieldProps & {
  value: string
  inputTypes?: InputType[]
}

export default function TelTextField(_props: TelTextFieldProps): JSX.Element {
  const { inputTypes = ['tel'], value = '', helperText = '', ...props } = _props
  const classes = useStyles()
  const [phoneNumber, setPhoneNumber] = useState('')

  const onlyTel = inputTypes.length === 1 && inputTypes[0] === 'tel'
  const onlySID = inputTypes.length === 1 && inputTypes[0] === 'sid'
  const isSID = inputTypes.includes('sid') && value.match(/^MG[a-zA-Z0-9]+$/)

  // debounce to set the phone number
  useEffect(() => {
    const t = setTimeout(() => {
      setPhoneNumber(value)
    }, DEBOUNCE_DELAY)
    return () => clearTimeout(t)
  }, [value])

  const skipValidatePhoneNumber = (): boolean => {
    if (!phoneNumber || props.disabled || !inputTypes.includes('tel')) {
      return true
    }
    if (onlyTel && !value.match(/(^\+)[0-9]+$/)) {
      return true
    }
    if (onlySID) {
      return true
    }
    if (isSID) {
      return true
    }
    return false
  }

  // check validation of the input phoneNumber through graphql
  const { data } = useQuery(isValidNumber, {
    pollInterval: 0,
    variables: { number: phoneNumber },
    fetchPolicy: 'cache-first',
    skip: skipValidatePhoneNumber(),
  })

  // fetch validation
  const valid = Boolean(data?.phoneNumberInfo?.valid)

  let adorn
  if (value === '' || isSID || props.disabled) {
    // no adornment if empty
  } else if (valid) {
    adorn = <Check className={classes.valid} />
  } else if (valid === false) {
    adorn = <Close className={classes.invalid} />
  }

  let iprops: Partial<InputProps>
  iprops = {}

  // if has inputProps from parent commponent, spread it in the iprops
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
    return props.onChange(e)
  }

  const getHelperText = (): TextFieldProps['helperText'] => {
    if (helperText) {
      return helperText
    }
    if (onlyTel) {
      return 'Please include a country code e.g. +1 (USA), +91 (India), +44 (UK)'
    }
    if (onlySID) {
      return ''
    }
    return 'For phone numbers, please include a country code e.g. +1 (USA), +91 (India), +44 (UK)'
  }

  return (
    <TextField
      fullWidth
      {...props}
      InputProps={iprops}
      onChange={handleChange}
      value={value}
      helperText={getHelperText()}
    />
  )
}
