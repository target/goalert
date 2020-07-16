import React, { useEffect, useState } from 'react'
import TextField from '@material-ui/core/TextField'
import { Check, Close } from '@material-ui/icons'
import _ from 'lodash-es'
import InputAdornment from '@material-ui/core/InputAdornment'
import { makeStyles } from '@material-ui/core'
import { DEBOUNCE_DELAY } from '../config'

import gql from 'graphql-tag'
import { useQuery } from '@apollo/react-hooks'

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

export default function TelTextField({ InputProps, ...props }) {
  const classes = useStyles()
  const [phoneNumber, setPhoneNumber] = useState(null)

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
    skip: !phoneNumber,
  })

  // fetch validation
  const valid = _.get(data, 'phoneNumberInfo.valid', null)

  let adorn
  if (valid) {
    adorn = <Check className={classes.valid} />
  } else if (props.error) {
    adorn = <Close className={classes.invalid} />
  }

  // add live validation icon to the right of the textfield as an endAdornment
  let iprops = InputProps
  if (adorn) {
    iprops = {
      endAdornment: <InputAdornment position='end'>{adorn}</InputAdornment>,
      ...InputProps,
    }
  }

  return (
    <div>
      <TextField {...props} InputProps={iprops} />
    </div>
  )
}
