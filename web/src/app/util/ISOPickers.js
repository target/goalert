import React, { useState, useEffect } from 'react'
import {
  KeyboardDatePicker,
  KeyboardDateTimePicker,
  KeyboardTimePicker,
} from '@material-ui/pickers'
import { useSelector } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { DateTime } from 'luxon'
import { getPaddedLocaleFormatString, getFormatMask } from './timeFormat'

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
const timeMask = getFormatMask(timeFormat)
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
      mask={timeMask}
      value={inputValue}
      onChange={handleInputChange}
      {...otherProps}
    />
  )
}

const dateTimeFormat = getPaddedLocaleFormatString(DateTime.DATETIME_SHORT)
const dateTimeMask = getFormatMask(dateTimeFormat)
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
      mask={dateTimeMask}
      value={inputValue}
      onChange={handleInputChange}
      {...otherProps}
    />
  )
}

const dateFormat = getPaddedLocaleFormatString(DateTime.DATE_SHORT)
const dateMask = getFormatMask(dateFormat)
export function ISODatePicker(props) {
  const { value, onChange, ...otherProps } = props
  const [inputValue, handleInputChange] = useDatePicker(value, onChange)

  return (
    <KeyboardDatePicker
      format={dateFormat}
      mask={dateMask}
      value={inputValue}
      onChange={handleInputChange}
      {...otherProps}
    />
  )
}
