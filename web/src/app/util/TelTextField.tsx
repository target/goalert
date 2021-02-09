import React, { useEffect, useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import TextField, { TextFieldProps } from '@material-ui/core/TextField'
import { InputProps } from '@material-ui/core/Input'
import { Check, Close } from '@material-ui/icons'
import _ from 'lodash'
import InputAdornment from '@material-ui/core/InputAdornment'
import { makeStyles } from '@material-ui/core'
import { DEBOUNCE_DELAY } from '../config'

const isValidNumber = gql`
  query($number: String!) {
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
  const { data } = useQuery(isValidNumber, {
    pollInterval: 0,
    variables: { number: '+' + phoneNumber },
    fetchPolicy: 'cache-first',
    skip: !phoneNumber || props.disabled,
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
    e.target.value = '+' + e.target.value.replace(/[^0-9]/g, '')
    return props.onChange(e)
  }

  return (
    <TextField
      fullWidth
      {...props}
      InputProps={iprops}
      type={props.type || 'tel'}
      onChange={handleChange}
      value={(props.value || '').replace(/[^0-9]/g, '')}
    />
  )
}
