import React, { useState, useEffect } from 'react'
import { DateTime, DateTimeUnit } from 'luxon'
import { TextField, TextFieldProps, useTheme } from '@mui/material'
import { useURLParam } from '../actions'
import { FormContainerContext } from '../forms/context'

interface ISOPickerProps extends ISOTextFieldProps {
  format: string
  timeZone?: string
  truncateTo: DateTimeUnit
  type: 'time' | 'date' | 'datetime-local'

  min?: string
  max?: string

  // softMin and softMax are used to set the min and max values of the input
  // without restricting the value that can be entered.
  //
  // Values outside of the softMin and softMax will be stored in local state
  // and will not be passed to the parent onChange.
  softMin?: string
  softMinLabel?: string
  softMax?: string
  softMaxLabel?: string
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

    softMin,
    softMax,

    softMinLabel,
    softMaxLabel,

    ...textFieldProps
  } = props

  const theme = useTheme()
  const [_zone] = useURLParam('tz', 'local')
  const zone = timeZone || _zone
  const valueAsDT = React.useMemo(
    () => (props.value ? DateTime.fromISO(props.value, { zone }) : null),
    [props.value, zone],
  )

  // store input value as DT.format() string. pass to parent onChange as ISO string
  const [inputValue, setInputValue] = useState(
    valueAsDT?.toFormat(format) ?? '',
  )

  function getSoftValidationError(value: string): string {
    if (props.disabled) return ''
    let dt: DateTime
    try {
      dt = DateTime.fromISO(value, { zone })
    } catch (e) {
      return `Invalid date/time`
    }

    if (softMin) {
      const sMin = DateTime.fromISO(softMin)
      if (dt < sMin) {
        return `Value must be after ${softMinLabel || sMin.toFormat(format)}`
      }
    }
    if (softMax) {
      const sMax = DateTime.fromISO(softMax)
      if (dt > sMax) {
        return `Value must be before ${softMaxLabel || sMax.toFormat(format)}`
      }
    }

    return ''
  }

  const { setValidationError } = React.useContext(FormContainerContext) as {
    setValidationError: (name: string, errMsg: string) => void
  }
  useEffect(
    () =>
      setValidationError(props.name || '', getSoftValidationError(inputValue)),
    [inputValue, props.disabled, valueAsDT, props.name, softMin, softMax],
  )

  useEffect(() => {
    setInputValue(valueAsDT?.toFormat(format) ?? '')
  }, [valueAsDT])

  const dtToISO = (dt: DateTime): string => {
    return dt.startOf(truncateTo).toUTC().toISO()
  }

  // parseInputToISO takes input from the form control and returns a string
  // ISO value representing the current form value ('' if invalid or empty).
  function parseInputToISO(input?: string): string {
    if (!input) return ''

    // handle input in specific format e.g. MM/dd/yyyy
    try {
      const inputAsDT = DateTime.fromFormat(input, format, { zone })

      if (valueAsDT && type === 'time') {
        return dtToISO(
          valueAsDT.set({
            hour: inputAsDT.hour,
            minute: inputAsDT.minute,
          }),
        )
      }
      return dtToISO(inputAsDT)
    } catch (e) {
      // ignore if input doesn't match format
    }

    // if format string invalid, try validating input as iso string
    try {
      const iso = DateTime.fromISO(input, { zone })
      return dtToISO(iso)
    } catch (e) {
      // ignore if input doesn't match iso format
    }

    return ''
  }

  const handleChange = (newInputValue: string): void => {
    const newVal = parseInputToISO(newInputValue)
    if (!newVal) return
    if (getSoftValidationError(newVal)) return

    onChange(newVal)
  }

  useEffect(() => {
    handleChange(inputValue)
  }, [softMin, softMax]) // inputValue intentionally omitted to prevent loop

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
    setInputValue(e.target.value)
    handleChange(e.target.value)
  }

  const defaultLabel = type === 'time' ? 'Select a time...' : 'Select a date...'

  return (
    <TextField
      label={defaultLabel}
      {...textFieldProps}
      type={type}
      value={inputValue}
      onChange={handleInputChange}
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
