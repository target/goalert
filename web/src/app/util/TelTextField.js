import React from 'react'
import TextField from '@material-ui/core/TextField'
import { Check, Close } from '@material-ui/icons'
import _ from 'lodash-es'
import InputAdornment from '@material-ui/core/InputAdornment'

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

export default function TelTextField({ InputProps, ...props }) {
  const { value, error } = props
  const phoneNumber = '+' + value

  // check validation of the input phoneNumber through graphql
  const { data } = useQuery(isValidNumber, {
    // for some reason I need to add this to avoid wrong data immidiately after true
    pollInterval: 1000,
    variables: { number: phoneNumber },
  })

  console.log(phoneNumber)

  // fetch validation
  const valid = _.get(data, 'phoneNumberInfo.valid', null)
  console.log('Valid: ' + valid)

  // use error to detect if user click submit,
  // if error is true, meaning user clicked submit and has error
  console.log('error: ' + error)

  // add live validation icon to the right of the textfield as an endAdornment
  const iprops = Object.assign(
    {
      endAdornment: (
        <InputAdornment position='end'>
          {valid ? (
            <Check style={{ fill: 'green' }} />
          ) : error ? (
            <Close style={{ fill: 'red' }} />
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
