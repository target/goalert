import React, { useState, useEffect } from 'react'
import { DateTime, DateTimeUnit } from 'luxon'
import { TextField, TextFieldProps, useTheme } from '@mui/material'
import { useURLParam } from '../actions'

interface ISOPickerProps extends ISOTextFieldProps {
  format: string
  timeZone?: string
  truncateTo: DateTimeUnit
  type: 'time' | 'date' | 'datetime-local'

  min?: string
  max?: string
}

// Used for the native textfield component or the nested input component
// that the Fallback renders.
type ISOTextFieldProps = Partial<Omit<TextFieldProps, 'value' | 'onChange'>> & {
  value?: string
  onChange: (value: string) => void
}

function ISOPicker(props: ISOPickerProps): JSX.Element {
  const {
    format,
    timeZone,
    truncateTo,
    type,

    value,
    onChange,
    min,
    max,

    ...textFieldProps
  } = props

  const theme = useTheme()
  const [_zone] = useURLParam('tz', 'local')
  const zone = timeZone || _zone
  let valueAsDT = props.value ? DateTime.fromISO(props.value, { zone }) : null

  // store input value as DT.format() string. pass to parent onChange as ISO string
  const [inputValue, setInputValue] = useState(
    valueAsDT?.toFormat(format) ?? '',
  )

  useEffect(() => {
    setInputValue(valueAsDT?.toFormat(format) ?? '')
  }, [valueAsDT])

  // update isopickers render on reset
  useEffect(() => {
    valueAsDT = props.value ? DateTime.fromISO(props.value, { zone }) : null
    setInputValue(valueAsDT ? valueAsDT.toFormat(format) : '')
  }, [props.value])

  const dtToISO = (dt: DateTime): string => {
    return dt.startOf(truncateTo).setZone(zone).toISO()
  }

  // parseInputToISO takes input from the form control and returns a string
  // ISO value representing the current form value ('' if invalid or empty).
  function parseInputToISO(input?: string): string {
    if (!input) return ''

    // handle input in specific format e.g. MM/dd/yyyy
    const inputAsDT = DateTime.fromFormat(input, format, { zone })
    if (inputAsDT.isValid) {
      if (valueAsDT && type === 'time') {
        return dtToISO(
          valueAsDT.set({
            hour: inputAsDT.hour,
            minute: inputAsDT.minute,
          }),
        )
      }
      return dtToISO(inputAsDT)
    }

    // if format string invalid, try validating input as iso string
    const iso = DateTime.fromISO(input, { zone })
    if (iso.isValid) return dtToISO(iso)

    return ''
  }

  function handleChange(e: React.ChangeEvent<HTMLInputElement>): void {
    setInputValue(e.target.value)
    const newVal = parseInputToISO(e.target.value)

    // only fire the parent's `onChange` handler when we have a new valid value,
    // taking care to ensure we ignore any zonal differences.
    if (!valueAsDT || (newVal && newVal !== valueAsDT.toISO())) {
      onChange(newVal)
    }
  }

  const defaultLabel = type === 'time' ? 'Select a time...' : 'Select a date...'

  return (
    <TextField
      label={defaultLabel}
      {...textFieldProps}
      type={type}
      value={valueAsDT ? valueAsDT.toFormat(format) : inputValue}
      onChange={handleChange}
      InputLabelProps={{ shrink: true, ...textFieldProps?.InputLabelProps }}
      inputProps={{
        min: min ? DateTime.fromISO(min, { zone }).toFormat(format) : undefined,
        max: max ? DateTime.fromISO(max, { zone }).toFormat(format) : undefined,
        style: {
          colorScheme: theme.palette.mode,
        },
        ...textFieldProps?.inputProps,
      }}
    />
  )
}

export function ISOTimePicker(props: ISOTextFieldProps): JSX.Element {
  return <ISOPicker {...props} format='HH:mm' truncateTo='minute' type='time' />
}

export function ISODatePicker(props: ISOTextFieldProps): JSX.Element {
  return (
    <ISOPicker {...props} format='yyyy-MM-dd' truncateTo='day' type='date' />
  )
}

export function ISODateTimePicker(props: ISOTextFieldProps): JSX.Element {
  return (
    <ISOPicker
      {...props}
      format={`yyyy-MM-dd'T'HH:mm`}
      truncateTo='minute'
      type='datetime-local'
    />
  )
}
