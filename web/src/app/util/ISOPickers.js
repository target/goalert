import React, { useState, useEffect } from 'react'
import {
  DatePicker,
  TimePicker,
  DateTimePicker,
  KeyboardDateTimePicker,
} from '@material-ui/pickers'
import { useSelector } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { DateTime } from 'luxon'
import { TextField, InputAdornment, IconButton } from '@material-ui/core'

import { inputtypes } from 'modernizr-esm/feature/inputtypes'
import { DateRange, AccessTime } from '@material-ui/icons'

function hasInputSupport(name) {
  if (new URLSearchParams(location.search).get('nativeInput') === '0') {
    return false
  }

  return inputtypes[name]
}

function useISOPicker(
  { value, onChange, timeZone, min, max, ...otherProps },
  { format, truncateTo, type, Fallback },
) {
  const native = hasInputSupport(type)
  const params = useSelector(urlParamSelector)
  const zone = timeZone || params('tz', 'local')
  const dtValue = value ? DateTime.fromISO(value, { zone }) : null
  const [inputValue, setInputValue] = useState(
    value && dtValue ? dtValue.toFormat(format) : '',
  )

  // parseInput takes input from the form control and returns a DateTime
  // object representing the value, or null (if invalid or empty).
  const parseInput = (input) => {
    if (input instanceof DateTime) return input
    if (!input) return null

    const dt = DateTime.fromFormat(input, format, { zone })
    if (dt.isValid) {
      if (dtValue && type === 'time') {
        return dtValue.set({
          hour: dt.hour,
          minute: dt.minute,
        })
      }
      return dt
    }

    const iso = DateTime.fromISO(input)
    if (iso.isValid) return iso

    return null
  }

  // inputToISO returns a UTC ISO timestamp representing the provided
  // input value, or an empty string if invalid.
  const inputToISO = (input) => {
    const val = parseInput(input)
    return val ? val.startOf(truncateTo).toUTC().toISO() : ''
  }

  useEffect(() => {
    setInputValue(value && dtValue ? dtValue.toFormat(format) : '')
  }, [value, zone])

  const handleChange = (e) => {
    setInputValue(e.target.value)

    const newVal = inputToISO(e.target.value)
    // Only fire the parent's `onChange` handler when we have a new valid value,
    // taking care to ensure we ignore any zonal differences.
    if (!dtValue || (newVal && newVal !== dtValue.toUTC().toISO())) {
      onChange(newVal)
    }
  }

  // shrink: true sets the label above the textfield so the placeholder can be properly seen
  const inputLabelProps = otherProps?.InputLabelProps ?? {}
  inputLabelProps.shrink = true

  // sets min and max if set
  const inputProps = otherProps?.inputProps ?? {}
  if (min) inputProps.min = DateTime.fromISO(min).toFormat(format)
  if (max) inputProps.max = DateTime.fromISO(max).toFormat(format)

  if (native) {
    return (
      <TextField
        type={type}
        value={inputValue}
        onChange={handleChange}
        {...otherProps}
        InputLabelProps={inputLabelProps}
        inputProps={inputProps}
      />
    )
  }

  const [emptyLabel, extraProps, FallbackIcon] =
    type === 'time'
      ? ['Select a time...', {}, AccessTime]
      : [
          'Select a date...',
          {
            leftArrowButtonProps: { 'data-cy': 'month-back' },
            rightArrowButtonProps: { 'data-cy': 'month-next' },
          },
          DateRange,
        ]

  return (
    <Fallback
      value={value ? dtValue : null}
      onChange={(v) => handleChange({ target: { value: v } })}
      showTodayButton
      minDate={min}
      maxDate={max}
      emptyLabel={emptyLabel}
      DialogProps={{
        'data-cy': 'picker-fallback',
      }}
      InputProps={{
        'data-cy-fallback-type': type,
        endAdornment: (
          <InputAdornment position='end'>
            <IconButton>
              <FallbackIcon />
            </IconButton>
          </InputAdornment>
        ),
      }}
      inputProps={inputProps}
      {...extraProps}
      {...otherProps}
    />
  )
}

export function ISOTimePicker(props) {
  return useISOPicker(props, {
    format: 'HH:mm',
    truncateTo: 'minute',
    type: 'time',
    Fallback: TimePicker,
  })
}

export function ISODateTimePicker(props) {
  return useISOPicker(props, {
    format: `yyyy-MM-dd'T'HH:mm`,
    Fallback: DateTimePicker,
    truncateTo: 'minute',
    type: 'datetime-local',
  })
}

export function ISOKeyboardDateTimePicker(props) {
  return useISOPicker(props, {
    format: `yyyy-MM-dd'T'HH:mm`,
    Fallback: KeyboardDateTimePicker,
    truncateTo: 'minute',
    type: 'datetime-local',
  })
}

export function ISODatePicker(props) {
  return useISOPicker(props, {
    format: 'yyyy-MM-dd',
    Fallback: DatePicker,
    truncateTo: 'day',
    type: 'date',
  })
}
