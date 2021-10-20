import React, { useState, useEffect } from 'react'
import { useSelector } from 'react-redux'
import { DateTime, DurationObjectUnits } from 'luxon'
import { TextField, OutlinedTextFieldProps } from '@material-ui/core'
import V5TextField from '@mui/material/TextField'
import DatePicker from '@mui/lab/DatePicker'
import DateTimePicker from '@mui/lab/DateTimePicker'
import TimePicker from '@mui/lab/TimePicker'
import { inputtypes } from 'modernizr-esm/feature/inputtypes'
import { urlParamSelector } from '../selectors'

interface ISOPickerProps extends PickerProps {
  Fallback: typeof TimePicker | typeof DatePicker | typeof DateTimePicker
  format: string
  timeZone?: string
  truncateTo: keyof DurationObjectUnits
  type: 'time' | 'day' | 'datetime-local'
}

type PickerProps = Partial<Omit<OutlinedTextFieldProps, 'value'>> & {
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
  const { Fallback, format, onChange, timeZone, truncateTo, type, ...rest } =
    props

  const native = hasInputSupport(type)
  const getURLParam = useSelector(urlParamSelector)
  const zone = timeZone || (getURLParam('tz', 'local') as string)
  const valueAsDT = props.value ? DateTime.fromISO(props.value, { zone }) : null
  const [inputValue, setInputValue] = useState(
    valueAsDT ? valueAsDT.toFormat(format) : '',
  )

  const dtToISO = (dt: DateTime): string => {
    return dt.startOf(truncateTo).toUTC().toISO()
  }

  // parseInputToISO takes input from the form control and returns a string
  // ISO value representing the current form value ('' if invalid or empty).
  function parseInputToISO(input?: string | DateTime): string {
    if (!input) return ''
    if (input instanceof DateTime) return dtToISO(input)

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
    const iso = DateTime.fromISO(input)
    if (iso.isValid) return dtToISO(iso)

    return ''
  }

  useEffect(() => {
    setInputValue(props.value && valueAsDT ? valueAsDT.toFormat(format) : '')
  }, [props.value, zone])

  function handleNativeChange(e: React.ChangeEvent<HTMLInputElement>): void {
    setInputValue(e.target.value)
    const newVal = parseInputToISO(e.target.value)

    // Only fire the parent's `onChange` handler when we have a new valid value,
    // taking care to ensure we ignore any zonal differences.
    if (!valueAsDT || (newVal && newVal !== valueAsDT.toUTC().toISO())) {
      onChange(newVal)
    }
  }

  function handleFallbackChange(
    date: DateTime | null,
    keyboardInputValue = '',
  ): void {
    // attempt to set value from DateTime object first
    if (date && date.isValid) {
      onChange(dtToISO(date))
      setInputValue(date.toFormat(format))
    } else {
      setInputValue(keyboardInputValue)
      const dt = DateTime.fromFormat(keyboardInputValue, format)
      if (dt.isValid) onChange(dtToISO(dt))
    }
  }

  if (native) {
    return (
      <TextField
        type={type}
        value={inputValue}
        onChange={handleNativeChange}
        {...rest}
      />
    )
  }

  return (
    <Fallback
      value={props.value ? valueAsDT : null}
      onChange={handleFallbackChange}
      showTodayButton
      minDate={props?.inputProps?.min}
      maxDate={props?.inputProps?.max}
      label={type === 'time' ? 'Select a time...' : 'Select a date...'}
      renderInput={(params) => <V5TextField {...params} />}
    />
  )
}

export function ISOTimePicker(props: PickerProps): JSX.Element {
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

export function ISODatePicker(props: PickerProps): JSX.Element {
  return (
    <ISOPicker
      {...props}
      format='yyyy-MM-dd'
      truncateTo='day'
      type='day'
      Fallback={DatePicker}
    />
  )
}

export function ISODateTimePicker(props: PickerProps): JSX.Element {
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
