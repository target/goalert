import React, { useState, useEffect } from 'react'
import { DatePicker, TimePicker, DateTimePicker } from '@material-ui/pickers'
import { useSelector } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { DateTime } from 'luxon'
import { TextField, InputAdornment, IconButton } from '@material-ui/core'

import Modernizr from '../../modernizr.config'
import { DateRange, AccessTime } from '@material-ui/icons'

function hasInputSupport(name) {
  if (new URLSearchParams(location.search).get('nativeInput') === '0') {
    return false
  }

  return Modernizr.inputtypes[name]
}

function useISOPicker(
  { value, onChange, timeZone, ...otherProps },
  { format, truncateTo, type, Fallback },
) {
  const native = hasInputSupport(type)
  const params = useSelector(urlParamSelector)
  const zone = timeZone || params('tz', 'local')
  const dtValue = value ? DateTime.fromISO(value, { zone }) : null
  const [inputValue, setInputValue] = useState(dtValue?.toFormat(format) ?? '')

  // parseInput takes input from the form control and returns a DateTime
  // object representing the value, or null (if invalid or empty).
  const parseInput = (input) => {
    if (input instanceof DateTime) return input
    if (!input) return null

    const dt = DateTime.fromFormat(input, format, { zone })
    if (dt.isValid) {
      if (type === 'time') {
        return DateTime.fromISO(value, { zone }).set({
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
    setInputValue(dtValue.toFormat(format))
  }, [value, zone])

  const handleChange = (e) => {
    setInputValue(e.target.value)

    const newVal = inputToISO(e.target.value)
    // Only fire the parent's `onChange` handler when we have a new valid value,
    // taking care to ensure we ignore any zonal differences.
    if (newVal && newVal !== dtValue?.toUTC().toISO()) {
      onChange(newVal)
    }
  }

  if (native) {
    return (
      <TextField
        type={type}
        value={inputValue}
        onChange={handleChange}
        {...otherProps}
      />
    )
  }

  const extraProps = {}
  if (type !== 'time') {
    extraProps.leftArrowButtonProps = { 'data-cy': 'month-back' }
    extraProps.rightArrowButtonProps = { 'data-cy': 'month-next' }
  }

  const FallbackIcon = type === 'time' ? AccessTime : DateRange
  return (
    <Fallback
      value={dtValue}
      onChange={(v) => handleChange({ target: { value: v } })}
      showTodayButton
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

export function ISODatePicker(props) {
  return useISOPicker(props, {
    format: 'yyyy-MM-dd',
    Fallback: DatePicker,
    truncateTo: 'day',
    type: 'date',
  })
}
