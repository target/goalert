import React, { useState, useEffect } from 'react'
import { DatePicker, TimePicker, DateTimePicker } from '@material-ui/pickers'
import { useSelector } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { DateTime } from 'luxon'
import { TextField, InputAdornment, IconButton } from '@material-ui/core'

import Modernizr from '../../modernizr.config'
import { DateRange } from '@material-ui/icons'

function hasInputSupport(name) {
  return Modernizr.inputtypes[name]
}

function useISOPicker(
  { value, onChange, ...otherProps },
  { format, truncateTo, type, Fallback },
) {
  const native = hasInputSupport(type)
  const params = useSelector(urlParamSelector)
  const zone = params('tz', 'local')
  const dtValue = DateTime.fromISO(value, { zone })
  const [inputValue, setInputValue] = useState(dtValue.toFormat(format))

  const parseInput = input => {
    const iso = DateTime.fromISO(input)
    if (iso.isValid) return iso

    const dt = DateTime.fromFormat(input, format, { zone })
    if (dt.isValid) {
      if (type === 'time') {
        return dtValue.set({
          hour: dt.hour,
          minute: dt.minute,
        })
      }
      return dt
    }

    return null
  }
  const isoInput = input => {
    const val = parseInput(input)
    return val
      ? val
          .startOf(truncateTo)
          .toUTC()
          .toISO()
      : ''
  }

  useEffect(() => {
    setInputValue(dtValue.toFormat(format))
  }, [value, zone])

  const handleChange = e => {
    setInputValue(e.target.value)

    const newVal = isoInput(e.target.value)
    if (newVal && newVal !== value) {
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

  return (
    <Fallback
      value={dtValue}
      onChange={v => handleChange({ target: { value: v } })}
      showTodayButton
      autoOk
      InputProps={{
        endAdornment: (
          <InputAdornment position='end'>
            <IconButton>
              <DateRange />
            </IconButton>
          </InputAdornment>
        ),
      }}
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
