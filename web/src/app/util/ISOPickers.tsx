import React, { useState, useEffect } from 'react'
import { DateTime, DateTimeUnit } from 'luxon'
import { TextField, TextFieldProps } from '@mui/material'
import DatePicker from '@mui/lab/DatePicker'
import DateTimePicker from '@mui/lab/DateTimePicker'
import TimePicker from '@mui/lab/TimePicker'
import { inputtypes } from 'modernizr-esm/feature/inputtypes'
import { useURLParam } from '../actions'

interface ISOPickerProps extends ISOTextFieldProps {
  Fallback: typeof TimePicker | typeof DatePicker | typeof DateTimePicker
  format: string
  timeZone?: string
  truncateTo: DateTimeUnit
  type: 'time' | 'date' | 'datetime-local'

  min?: string
  max?: string
}

// Used for the native textfield component or the nested input component
// that the Fallback renders.
type ISOTextFieldProps = Partial<Omit<TextFieldProps, 'value'>> & {
  value?: string
  onChange: (value: string) => void
}

function hasInputSupport(name: string): boolean {
  if (new URLSearchParams(location.search).get('nativeInput') === '0') {
    return false
  }

  return inputtypes[name]
}

function ISOPicker(props: ISOPickerProps): JSX.Element {
  const {
    Fallback,
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

  const native = hasInputSupport(type)
  const [_zone] = useURLParam('tz', 'local')
  const zone = timeZone || _zone
  const valueAsDT = props.value ? DateTime.fromISO(props.value, { zone }) : null

  // store input value as DT.format() string. pass to parent onChange as ISO string
  const [inputValue, setInputValue] = useState(
    valueAsDT?.toFormat(format) ?? '',
  )
  useEffect(() => {
    setInputValue(valueAsDT?.toFormat(format) ?? '')
  }, [valueAsDT])

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

  function handleNativeChange(e: React.ChangeEvent<HTMLInputElement>): void {
    setInputValue(e.target.value)
    const newVal = parseInputToISO(e.target.value)

    // only fire the parent's `onChange` handler when we have a new valid value,
    // taking care to ensure we ignore any zonal differences.
    if (!valueAsDT || (newVal && newVal !== valueAsDT.toISO())) {
      onChange(newVal)
    }
  }

  function handleFallbackChange(
    date: DateTime | null,
    keyboardInputValue = '',
  ): void {
    // attempt to set value from DateTime object first
    if (date?.isValid) {
      setInputValue(date.toFormat(format))
      onChange(dtToISO(date))
    } else {
      setInputValue(keyboardInputValue)
      // likely invalid, but validate keyboard input just to be sure
      const dt = DateTime.fromFormat(keyboardInputValue, format, { zone })
      if (dt.isValid) onChange(dtToISO(dt))
      else onChange(keyboardInputValue) // set invalid input for form validation
    }
  }

  const defaultLabel = type === 'time' ? 'Select a time...' : 'Select a date...'
  if (native) {
    return (
      <TextField
        label={defaultLabel}
        {...textFieldProps}
        type={type}
        value={valueAsDT ? valueAsDT.toFormat(format) : inputValue}
        onChange={handleNativeChange}
        InputLabelProps={{ shrink: true, ...textFieldProps?.InputLabelProps }}
        inputProps={{
          min: min
            ? DateTime.fromISO(min, { zone }).toFormat(format)
            : undefined,
          max: max
            ? DateTime.fromISO(max, { zone }).toFormat(format)
            : undefined,
          ...textFieldProps?.inputProps,
        }}
      />
    )
  }

  return (
    <Fallback
      value={valueAsDT}
      onChange={handleFallbackChange}
      showTodayButton
      minDate={min ? DateTime.fromISO(min, { zone }) : undefined}
      maxDate={max ? DateTime.fromISO(max, { zone }) : undefined}
      disabled={textFieldProps?.disabled}
      renderInput={(params) => (
        <TextField
          data-cy-fallback-type={type}
          {...params}
          label={defaultLabel}
          {...textFieldProps}
        />
      )}
      PopperProps={{
        // @ts-expect-error DOM attribute for testing
        'data-cy': props.name + '-picker-fallback',
      }}
      style={{ width: 'fit-container' }}
    />
  )
}

export function ISOTimePicker(props: ISOTextFieldProps): JSX.Element {
  return (
    <ISOPicker
      {...props}
      format='HH:mm'
      truncateTo='minute'
      type='time'
      Fallback={TimePicker}
    />
  )
}

export function ISODatePicker(props: ISOTextFieldProps): JSX.Element {
  return (
    <ISOPicker
      {...props}
      format='yyyy-MM-dd'
      truncateTo='day'
      type='date'
      Fallback={DatePicker}
    />
  )
}

export function ISODateTimePicker(props: ISOTextFieldProps): JSX.Element {
  return (
    <ISOPicker
      {...props}
      format={`yyyy-MM-dd'T'HH:mm`}
      truncateTo='minute'
      type='datetime-local'
      Fallback={DateTimePicker}
    />
  )
}
