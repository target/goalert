import React, { useEffect, useState } from 'react'
import { useQuery, gql } from 'urql'
import TextField from '@mui/material/TextField'
import { InputProps } from '@mui/material/Input'
import { Check, Close } from '@mui/icons-material'
import InputAdornment from '@mui/material/InputAdornment'
import { DEBOUNCE_DELAY } from '../config'
import { DestinationFieldConfig, DestinationType } from '../../schema'
import AppLink from '../util/AppLink'
import { green, red } from '@mui/material/colors'

const isValidValue = gql`
  query ValidateDestination($input: DestinationFieldValidateInput!) {
    destinationFieldValidate(input: $input)
  }
`

const noSuspense = { suspense: false }

function trimPrefix(value: string, prefix: string): string {
  if (!prefix) return value
  if (!value) return value
  if (value.startsWith(prefix)) return value.slice(prefix.length)
  return value
}

export type DestinationInputDirectProps = DestinationFieldConfig & {
  value: string
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void
  destType: DestinationType

  disabled?: boolean
}

/**
 * DestinationInputDirect is a text field that allows the user to enter a
 * destination directly. It supports validation and live feedback.
 *
 * You should almost never use this component directly. Instead, use
 * DestinationField, which will select the correct component based on the
 * destination type.
 */
export default function DestinationInputDirect(
  props: DestinationInputDirectProps,
): JSX.Element {
  const [debouncedValue, setDebouncedValue] = useState(props.value)

  // debounce the input
  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(props.value)
    }, DEBOUNCE_DELAY)
    return () => {
      clearTimeout(handler)
    }
  }, [props.value])

  // check validation of the input phoneNumber through graphql
  const [{ data }] = useQuery<{ destinationFieldValidate: boolean }>({
    query: isValidValue,
    variables: {
      input: {
        destType: props.destType,
        value: debouncedValue,
        fieldID: props.fieldID,
      },
    },
    requestPolicy: 'cache-first',
    pause: !props.value || props.disabled || !props.supportsValidation,
    context: noSuspense,
  })

  // fetch validation
  const valid = !!data?.destinationFieldValidate

  let adorn
  if (!props.value || !props.supportsValidation) {
    // no adornment if empty
  } else if (valid) {
    adorn = <Check sx={{ fill: green[500] }} />
  } else if (valid === false) {
    adorn = <Close sx={{ fill: red[500] }} />
  }

  let iprops: Partial<InputProps> = {}

  if (props.prefix) {
    iprops.startAdornment = (
      <InputAdornment position='start' sx={{ mb: '0.1em' }}>
        {props.prefix}
      </InputAdornment>
    )
  }

  // add live validation icon to the right of the textfield as an endAdornment
  if (adorn && props.value === debouncedValue) {
    iprops = {
      endAdornment: <InputAdornment position='end'>{adorn}</InputAdornment>,
      ...iprops,
    }
  }

  // remove unwanted character
  function handleChange(e: React.ChangeEvent<HTMLInputElement>): void {
    if (!props.onChange) return
    if (!e.target.value) return props.onChange(e)

    if (props.inputType === 'tel') {
      e.target.value = e.target.value.replace(/[^0-9+]/g, '')
    }

    e.target.value = props.prefix + e.target.value
    return props.onChange(e)
  }

  return (
    <TextField
      fullWidth
      disabled={props.disabled}
      InputProps={iprops}
      type={props.inputType}
      placeholder={props.placeholderText}
      label={props.labelSingular}
      helperText={
        props.hintURL ? (
          <AppLink newTab to={props.hintURL}>
            {props.hint}
          </AppLink>
        ) : (
          props.hint
        )
      }
      onChange={handleChange}
      value={trimPrefix(props.value, props.prefix)}
    />
  )
}
