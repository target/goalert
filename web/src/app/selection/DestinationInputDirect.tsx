import React, { useEffect, useState } from 'react'
import { useQuery, gql } from 'urql'
import TextField from '@mui/material/TextField'
import { InputProps } from '@mui/material/Input'
import { Check, Close } from '@mui/icons-material'
import InputAdornment from '@mui/material/InputAdornment'
import makeStyles from '@mui/styles/makeStyles'
import { DEBOUNCE_DELAY } from '../config'
import { DestinationFieldConfig, DestinationType } from '../../schema'
import AppLink from '../util/AppLink'

const isValidValue = gql`
  query ValidateDestination($input: DestinationFieldValidateInput!) {
    destinationFieldValidate(input: $input)
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

function trimPrefix(value: string, prefix: string): string {
  if (!prefix) return value
  if (!value) return value
  if (value.startsWith(prefix)) return value.slice(prefix.length)
  return value
}

export type DestinationInputDirectProps = {
  value: string
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void
  config: DestinationFieldConfig
  destType: DestinationType

  disabled?: boolean
}

export default function DestinationInputDirect(
  props: DestinationInputDirectProps,
): JSX.Element {
  const classes = useStyles()

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
        fieldID: props.config.fieldID,
      },
    },
    requestPolicy: 'cache-first',
    pause: !props.value || props.disabled || !props.config.supportsValidation,
    context: noSuspense,
  })

  // fetch validation
  const valid = !!data?.destinationFieldValidate

  let adorn
  if (!props.value || !props.config.supportsValidation) {
    // no adornment if empty
  } else if (valid) {
    adorn = <Check className={classes.valid} />
  } else if (valid === false) {
    adorn = <Close className={classes.invalid} />
  }

  let iprops: Partial<InputProps> = {}

  if (props.config.prefix) {
    iprops.startAdornment = (
      <InputAdornment position='start' style={{ marginBottom: '0.1em' }}>
        {props.config.prefix}
      </InputAdornment>
    )
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

    if (props.config.inputType === 'tel') {
      e.target.value = e.target.value.replace(/[^0-9+]/g, '')
    }

    e.target.value = props.config.prefix + e.target.value
    return props.onChange(e)
  }

  // TODO: what to do with input limiting (e.g., only allow digits)
  return (
    <TextField
      fullWidth
      InputProps={iprops}
      type={props.config.inputType}
      placeholder={props.config.placeholderText}
      label={props.config.labelSingular}
      helperText={
        props.config.hintURL ? (
          <AppLink newTab to={props.config.hintURL}>
            {props.config.hint}
          </AppLink>
        ) : (
          props.config.hint
        )
      }
      onChange={handleChange}
      value={trimPrefix(props.value, props.config.prefix)}
    />
  )
}
