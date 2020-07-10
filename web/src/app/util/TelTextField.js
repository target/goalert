import React from 'react'
import TextField from '@material-ui/core/TextField'
import { Check, Close } from '@material-ui/icons'
import _ from 'lodash-es'
import InputAdornment from '@material-ui/core/InputAdornment'
import { makeStyles } from '@material-ui/core'

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
  const { value, error } = props
  const phoneNumber = '+' + value

  // check validation of the input phoneNumber through graphql
  const { data } = useQuery(isValidNumber, {
    pollInterval: 1000,
    variables: { number: phoneNumber },
  })

  // fetch validation
  const valid = _.get(data, 'phoneNumberInfo.valid', null)

  // add live validation icon to the right of the textfield as an endAdornment
  const iprops = Object.assign(
    {
      endAdornment: (
        <InputAdornment position='end'>
          {valid ? (
            // change to makeStyle
            <Check className={classes.valid} />
          ) : error ? (
            <Close className={classes.invalid} />
          ) : (
            <div />
          )}
        </InputAdornment>
      ),
    },
    InputProps,
  )

  return (
    <div>
      <TextField {...props} InputProps={iprops} />
    </div>
  )
}
