import React, { useState, useEffect } from 'react'
import {
  KeyboardDatePicker,
  KeyboardDateTimePicker,
  KeyboardTimePicker,
} from '@material-ui/pickers'
import { useSelector } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { DateTime, FixedOffsetZone } from 'luxon'

const fixed = DateTime.fromObject({
  month: 1,
  day: 2,
  hour: 15, // 3pm
  minute: 4,
  second: 5,
  year: 2006,
  millisecond: 99,
}).setZone(FixedOffsetZone.instance(-7))

const localeKeys = [
  'yyyy',
  'yy',
  'y',

  'LL',
  'L',
  'LLLL',
  'LLL',
  'LLLLL',
  'dd',
  'd',
  'u',
  'S',

  'HH',
  'H',
  'hh',
  'h',
  'mm',
  'm',
  'ss',
  's',

  'ZZZZZ',
  'ZZZZ',
  'ZZZ',
  'ZZ',
  'Z',
  'z',
  'a',
  'ccc',
  'cccc',
  'ccccc',
]

export const getPaddedLocaleFormatString = opts => {
  let s = fixed.toLocaleString(opts)
  localeKeys.forEach(key => {
    s = s.replace(fixed.toFormat(key), key)
  })
  return (
    s
      // ensure we always use the padded versions
      .replace(/H+/, 'HH')
      .replace(/h+/, 'hh')
      .replace(/m+/, 'mm')
      .replace(/s+/, 'ss')
      .replace(/d+/, 'dd')
      .replace(/\bL\b/, 'LL')
      .replace(/\bM\b/, 'MM')
      .replace(/\by\b/, 'yy')
  )
}

function useDatePicker(value, onChange, startOf) {
  const params = useSelector(urlParamSelector)
  const zone = params('tz', 'local')
  const [dtVal, setDTVal] = useState(DateTime.fromISO(value, { zone }))

  useEffect(() => {
    setDTVal(DateTime.fromISO(value, { zone }))
  }, [value])
  useEffect(() => {
    setDTVal(dtVal.setZone(zone))
  }, [zone])

  return [
    dtVal,
    e => {
      setDTVal(e)
      if (e && e.isValid) onChange(e.startOf(startOf).toISO()) // only trigger external onChange when we have a valid value
    },
  ]
}

const timeFormat = getPaddedLocaleFormatString('t')
export function ISOTimePicker(props) {
  const { value, onChange, ...otherProps } = props
  const [inputValue, handleInputChange] = useDatePicker(
    value,
    onChange,
    'minute',
  )

  return (
    <KeyboardTimePicker
      format={timeFormat}
      mask={fixed.toFormat(timeFormat).replace(/[0-9APM]/g, '_')}
      value={inputValue}
      onChange={handleInputChange}
      {...otherProps}
    />
  )
}

const dateTimeFormat = getPaddedLocaleFormatString(DateTime.DATETIME_SHORT)
export function ISODateTimePicker(props) {
  const { value, onChange, ...otherProps } = props
  const [inputValue, handleInputChange] = useDatePicker(
    value,
    onChange,
    'minute',
  )

  return (
    <KeyboardDateTimePicker
      format={dateTimeFormat}
      mask={fixed.toFormat(dateTimeFormat).replace(/[0-9APM]/g, '_')}
      value={inputValue}
      onChange={handleInputChange}
      {...otherProps}
    />
  )
}

const dateFormat = getPaddedLocaleFormatString(DateTime.DATE_SHORT)
export function ISODatePicker(props) {
  const { value, onChange, ...otherProps } = props
  const [inputValue, handleInputChange] = useDatePicker(value, onChange)

  return (
    <KeyboardDatePicker
      format={dateFormat}
      mask={fixed.toFormat(dateFormat).replace(/[0-9APM]/g, '_')}
      value={inputValue}
      onChange={handleInputChange}
      {...otherProps}
    />
  )
}
