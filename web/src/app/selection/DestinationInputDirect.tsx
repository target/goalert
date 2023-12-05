import React, { useEffect, useState } from 'react'
import { useQuery, gql } from 'urql'
import TextField, { TextFieldProps } from '@mui/material/TextField'
import { InputProps } from '@mui/material/Input'
import { Check, Close } from '@mui/icons-material'
import InputAdornment from '@mui/material/InputAdornment'
import makeStyles from '@mui/styles/makeStyles'
import { DEBOUNCE_DELAY } from '../config'
import { InputFieldConfig } from '../../schema'
import AppLink from '../util/AppLink'

const isValidValue = gql`
  query ($type: InputFieldDataType!, $value: String!) {
    inputFieldValidate(type: $type, value: $value)
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

export default function DestinationField(
  props: TextFieldProps & { value: string; config: InputFieldConfig },
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
  const [{ data }] = useQuery({
    query: isValidValue,
    variables: { value: debouncedValue, type: props.config.dataType },
    requestPolicy: 'cache-first',
    pause: !props.value || props.disabled || !props.config.supportsValidation,
    context: noSuspense,
  })

  // fetch validation
  const valid = !!data?.inputFieldValidate

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

    e.target.value = props.config.prefix + e.target.value
    return props.onChange(e)
  }

  // TODO: what to do with input limiting (e.g., only allow digits)
  return (
    <TextField
      fullWidth
      {...props}
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
