import React, { useState, useEffect, ChangeEvent, FC, ReactNode } from 'react'
import {
  DatePicker,
  TimePicker,
  DateTimePicker,
  TimePickerProps,
  DatePickerProps,
  DateTimePickerProps,
} from '@material-ui/pickers'
import { useSelector } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { DateTime, DurationUnit } from 'luxon'
import { TextField, InputAdornment, IconButton } from '@material-ui/core'

import Modernizr from '../../modernizr.config'
import { DateRange, AccessTime } from '@material-ui/icons'

type ISOPickersProps = {
  value?: string
  onChange: (newValue: string) => void
  timeZone?: string
  min?: string // yyyy-MM-dd'T'HH:mm:ss
  max?: string // yyyy-MM-dd'T'HH:mm:ss
}

// Supported fallback types
type FallbackType =
  | FC<TimePickerProps>
  | FC<DatePickerProps>
  | FC<DateTimePickerProps>

// Static settings defined in the ISOPickers variations (date, time, etc)
type ISOPickersSettings = {
  format: string
  Fallback: FallbackType
  truncateTo: DurationUnit
  type: 'date' | 'time' | 'datetime-local'
}

type Input = DateTime | string | null

function hasInputSupport(name: string): boolean {
  if (new URLSearchParams(location.search).get('nativeInput') === '0') {
    return false
  }

  // @ts-ignore: types are generated at build
  return Modernizr.inputtypes[name]
}

function useISOPicker(
  { value = '', onChange, timeZone, min, max, ...otherProps }: ISOPickersProps,
  { format, truncateTo, type, Fallback }: ISOPickersSettings,
): ReactNode {
  const native = hasInputSupport(type)
  const params = useSelector(urlParamSelector)
  const zone = timeZone || (params('tz', 'local') as string)
  const dtValue = DateTime.fromISO(value, { zone })
  const [inputValue, setInputValue] = useState(
    value ? dtValue.toFormat(format) : '',
  )

  // parseInput takes input from the form control and returns a DateTime
  // object representing the value, or null (if invalid or empty).
  const parseInput = (input: Input): DateTime | null => {
    if (input instanceof DateTime) return input
    if (!input) return null

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

    const iso = DateTime.fromISO(input)
    if (iso.isValid) return iso

    return null
  }

  // inputToISO returns a UTC ISO timestamp representing the provided
  // input value, or an empty string if invalid.
  const inputToISO = (input: Input): string => {
    const val = parseInput(input)
    return val ? val.startOf(truncateTo).toUTC().toISO() : ''
  }

  useEffect(() => {
    setInputValue(value ? dtValue.toFormat(format) : '')
  }, [value, zone])

  const handleChange = (val: Input): void => {
    const newVal = inputToISO(val)
    // Only fire the parent's `onChange` handler when we have a new valid value,
    // taking care to ensure we ignore any zonal differences.
    if (newVal && newVal !== dtValue.toUTC().toISO()) {
      onChange(newVal)
    }
  }

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>): void => {
    const val = e.target.value
    setInputValue(val)
    handleChange(val)
  }

  if (native) {
    return (
      <TextField
        type={type}
        value={inputValue}
        onChange={handleInputChange}
        inputProps={{ min, max }}
        {...otherProps}
        InputLabelProps={{
          ...otherProps,
          shrink: true,
        }}
      />
    )
  }

  const cypressProps: { [key: string]: { 'data-cy'?: string } } = {}
  cypressProps.DialogProps = { 'data-cy': 'picker-fallback' }
  if (type !== 'time') {
    cypressProps.leftArrowButtonProps = { 'data-cy': 'month-back' }
    cypressProps.rightArrowButtonProps = { 'data-cy': 'month-next' }
  }

  const FallbackIcon = type === 'time' ? AccessTime : DateRange
  return (
    <Fallback
      value={dtValue}
      onChange={handleChange}
      showTodayButton
      minDate={min}
      maxDate={max}
      InputProps={{
        // @ts-ignore
        'data-cy-fallback-type': type,
        endAdornment: (
          <InputAdornment position='end'>
            <IconButton>
              <FallbackIcon />
            </IconButton>
          </InputAdornment>
        ),
      }}
      {...cypressProps}
      {...otherProps}
      InputLabelProps={{
        ...otherProps,
        shrink: true,
      }}
    />
  )
}

export function ISOTimePicker(props: ISOPickersProps): ReactNode {
  return useISOPicker(props, {
    Fallback: TimePicker,
    format: 'HH:mm',
    truncateTo: 'minute',
    type: 'time',
  })
}

export function ISODateTimePicker(props: ISOPickersProps): ReactNode {
  return useISOPicker(props, {
    Fallback: DateTimePicker,
    format: `yyyy-MM-dd'T'HH:mm`,
    truncateTo: 'minute',
    type: 'datetime-local',
  })
}

export function ISODatePicker(props: ISOPickersProps): ReactNode {
  return useISOPicker(props, {
    Fallback: DatePicker,
    format: 'yyyy-MM-dd',
    truncateTo: 'day',
    type: 'date',
  })
}
