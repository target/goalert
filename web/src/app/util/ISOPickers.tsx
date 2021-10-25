import React, { useState } from 'react'
import { useSelector } from 'react-redux'
import { DateTime, DurationObjectUnits } from 'luxon'
import { TextField, TextFieldProps } from '@material-ui/core'
import V5TextField from '@mui/material/TextField'
import DatePicker from '@mui/lab/DatePicker'
import DateTimePicker from '@mui/lab/DateTimePicker'
import TimePicker from '@mui/lab/TimePicker'
import { inputtypes } from 'modernizr-esm/feature/inputtypes'
import { urlParamSelector } from '../selectors'

interface ISOPickerProps extends NativeProps {
  Fallback: typeof TimePicker | typeof DatePicker | typeof DateTimePicker
  format: string
  timeZone?: string
  truncateTo: keyof DurationObjectUnits
  type: 'time' | 'date' | 'datetime-local'

  min?: string
  max?: string
}

type NativeProps = Partial<Omit<TextFieldProps, 'value'>> & {
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
  const getURLParam = useSelector(urlParamSelector)
  const zone = timeZone || (getURLParam('tz', 'local') as string)
  const valueAsDT = props.value ? DateTime.fromISO(props.value, { zone }) : null

  // store input value as DT.format() string. pass to parent onChange as ISO string
  const [inputValue, setInputValue] = useState(
    valueAsDT ? valueAsDT.toFormat(format) : '',
  )

  const dtToISO = (dt: DateTime): string => {
    return dt.startOf(truncateTo).toUTC().setZone(zone).toISO()
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
    if (!valueAsDT || (newVal && newVal !== valueAsDT.toUTC().toISO())) {
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

  const label = type === 'time' ? 'Select a time...' : 'Select a date...'
  let fmt: string
  if (type === 'time') fmt = 'HH:mm'
  else if (type === 'date') fmt = 'yyyy-MM-dd'
  else fmt = "yyyy-MM-dd'T'HH:mm"

  if (native) {
    return (
      <TextField
        {...textFieldProps}
        type={type}
        value={valueAsDT ? valueAsDT.toFormat(format) : inputValue}
        onChange={handleNativeChange}
        label={label}
        inputProps={{
          min: min ? DateTime.fromISO(min, { zone }).toFormat(fmt) : undefined,
          max: max ? DateTime.fromISO(max, { zone }).toFormat(fmt) : undefined,
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
      label={label}
      renderInput={(params) => (
        // @ts-expect-error potential type mismatches until fully on v5
        <V5TextField
          data-cy-fallback-type={type}
          {...textFieldProps}
          {...params}
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

export function ISOTimePicker(props: NativeProps): JSX.Element {
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

export function ISODatePicker(props: NativeProps): JSX.Element {
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

export function ISODateTimePicker(props: NativeProps): JSX.Element {
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
