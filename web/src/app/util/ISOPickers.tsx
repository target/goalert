import React, { useState, useEffect, FC } from 'react'
import {
  DatePicker,
  TimePicker,
  DateTimePicker,
  DateTimePickerProps,
  DatePickerProps,
  TimePickerProps,
} from '@material-ui/pickers'
import { DateTime, DurationObjectUnits } from 'luxon'
import {
  TextField,
  InputAdornment,
  IconButton,
  TextFieldProps,
} from '@material-ui/core'
import { inputtypes } from 'modernizr-esm/feature/inputtypes'
import { DateRange, AccessTime } from '@material-ui/icons'
import { useURLParam } from '../actions/hooks'

function hasInputSupport(name: keyof InputTypesBoolean): boolean {
  if (new URLSearchParams(location.search).get('nativeInput') === '0') {
    return false
  }

  return inputtypes[name]
}

interface ISOPickerOptions {
  format: string
  truncateTo: keyof DurationObjectUnits
  type: keyof InputTypesBoolean
  Fallback: FC<DatePickerProps> | FC<DateTimePickerProps> | FC<TimePickerProps>
}

type OtherProps = Partial<TextFieldProps & ISOPickerOptions['Fallback']>

type ISOPickerProps = {
  value: string
  timeZone?: string
  min?: string
  max?: string
  onChange: (v: string) => void
} & OtherProps

function useISOPicker(
  { value, onChange, timeZone, min, max, ...otherProps }: ISOPickerProps,
  { format, truncateTo, type, Fallback }: ISOPickerOptions,
): JSX.Element {
  const native = hasInputSupport(type)
  const [_zone] = useURLParam('tz', 'local') // hooks can't be called conditionally
  const zone = timeZone || _zone

  const dtValue = value ? DateTime.fromISO(value, { zone }) : null
  const [inputValue, setInputValue] = useState(
    value && dtValue ? dtValue.toFormat(format) : '',
  )

  // parseInput takes input from the form control and returns a DateTime
  // object representing the value, or null (if invalid or empty).
  const parseInput = (input: DateTime | string): DateTime | null => {
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

    const iso = DateTime.fromISO(input, { zone })
    if (iso.isValid) return iso

    return null
  }

  // inputToISO returns a UTC ISO timestamp representing the provided
  // input value, or an empty string if invalid.
  const inputToISO = (input: DateTime | string): string => {
    const val = parseInput(input)
    return val ? val.startOf(truncateTo).toUTC().toISO() : ''
  }

  useEffect(() => {
    setInputValue(value && dtValue ? dtValue.toFormat(format) : '')
  }, [value, zone])

  // sets min and max if set
  const inputProps = otherProps?.inputProps ?? {}
  if (min) inputProps.min = DateTime.fromISO(min, { zone }).toFormat(format)
  if (max) inputProps.max = DateTime.fromISO(max, { zone }).toFormat(format)

  // NOTE the input is either a traditional ChangeEvent
  // or an object with a MaterialUiPickersDate as its target value
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleChange = (e: any): void => {
    setInputValue(e.target.value)

    const newVal = inputToISO(e.target.value)
    if (min && DateTime.fromISO(newVal) < DateTime.fromISO(min)) return
    if (max && DateTime.fromISO(newVal) > DateTime.fromISO(max)) return

    // Only fire the parent's `onChange` handler when we have a new valid value,
    // taking care to ensure we ignore any zonal differences.
    if (!dtValue || (newVal && newVal !== dtValue.toUTC().toISO())) {
      onChange(newVal)
    }
  }

  const inputDT = DateTime.fromISO(inputToISO(inputValue), { zone })
  let isValid = inputDT.isValid
  if (min && isValid) {
    isValid = inputDT >= DateTime.fromISO(min)
  }
  if (max && isValid) {
    isValid = inputDT <= DateTime.fromISO(max)
  }

  // shrink: true sets the label above the textfield so the placeholder can be properly seen
  const inputLabelProps = otherProps?.InputLabelProps ?? {}
  inputLabelProps.shrink = true

  if (native) {
    return (
      <TextField
        type={type}
        value={inputValue}
        onChange={handleChange}
        {...otherProps}
        InputLabelProps={inputLabelProps}
        inputProps={inputProps}
        error={otherProps.error || !isValid}
        onBlur={() => {
          if (dtValue) setInputValue(dtValue.toFormat(format))
        }}
      />
    )
  }

  let emptyLabel = 'Select a time...'
  const extraProps = {}
  if (type !== 'time') {
    emptyLabel = 'Select a date...'
    // @ts-expect-error DOM attribute for testing
    extraProps.leftArrowButtonProps = { 'data-cy': 'month-back' }
    // @ts-expect-error DOM attribute for testing
    extraProps.rightArrowButtonProps = { 'data-cy': 'month-next' }
  }

  const FallbackIcon = type === 'time' ? AccessTime : DateRange
  return (
    <Fallback
      value={value ? dtValue : null}
      onChange={(v) => handleChange({ target: { value: v } })}
      showTodayButton
      minDate={min}
      maxDate={max}
      emptyLabel={emptyLabel}
      DialogProps={{
        // @ts-expect-error DOM attribute for testing
        'data-cy': 'picker-fallback',
      }}
      InputProps={{
        // @ts-expect-error DOM attribute for testing
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

export function ISOTimePicker(props: ISOPickerProps): JSX.Element {
  return useISOPicker(props, {
    format: 'HH:mm',
    truncateTo: 'minute',
    type: 'time',
    Fallback: TimePicker,
  })
}

export function ISODateTimePicker(props: ISOPickerProps): JSX.Element {
  return useISOPicker(props, {
    format: `yyyy-MM-dd'T'HH:mm`,
    Fallback: DateTimePicker,
    truncateTo: 'minute',
    type: 'datetime-local',
  })
}

export function ISODatePicker(props: ISOPickerProps): JSX.Element {
  return useISOPicker(props, {
    format: 'yyyy-MM-dd',
    Fallback: DatePicker,
    truncateTo: 'day',
    type: 'date',
  })
}
