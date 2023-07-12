import React, { useState } from 'react'
import NumberField from './NumberField'
import { Grid, MenuItem, Select } from '@mui/material'
import { Duration, DurationLikeObject } from 'luxon'

export type DurationFieldProps = {
  value: string
  name: string
  label: string
  onChange: (newValue: string) => void
}

type Unit = 'minute' | 'hour' | 'day' | 'week'

// getDefaultUnit returns largest unit that can be used for the given ISO duration
// as an integer. For example, if the value is 120, the default unit is hour, but
// if the value is 121, the default unit is minute.
//
// If the value is empty, the default unit is minute.
function getDefaultUnit(value: string): Unit {
  if (!value) return 'minute'
  const dur = Duration.fromISO(value)
  if (dur.as('hours') / 24 / 7 === Math.floor(dur.as('hours') / 24 / 7))
    return 'week'
  if (dur.as('hours') / 24 === Math.floor(dur.as('hours') / 24)) return 'day'
  if (dur.as('hours') === Math.floor(dur.as('hours'))) return 'hour'
  return 'minute'
}

const mult = {
  minute: 1,
  hour: 60,
  day: 60 * 24,
  week: 60 * 24 * 7,
}

export const DurationField: React.FC<DurationFieldProps> = (props) => {
  const [unit, setUnit] = useState(getDefaultUnit(props.value))
  const val = Duration.fromISO(props.value).as('minute') / mult[unit]

  const handleChange = (val: number, u: Unit = unit): void => {
    const dur = Duration.fromObject({
      minutes: val * mult[u],
    } as DurationLikeObject)
    props.onChange(dur.toISO())
  }

  return (
    <Grid container sx={{ width: '100%' }}>
      <Grid item xs={8}>
        <NumberField
          fullWidth
          value={val.toString()}
          name={props.name}
          label={props.label}
          onChange={(e) => handleChange(parseInt(e.target.value, 10))}
          sx={{
            '& fieldset': {
              borderTopRightRadius: 0,
              borderBottomRightRadius: 0,
            },
          }}
        />
      </Grid>
      <Grid item xs={4}>
        <Select
          value={unit}
          onChange={(e) => {
            setUnit(e.target.value as Unit)
            handleChange(val, e.target.value as Unit)
          }}
          sx={{
            width: '100%',
            '& fieldset': {
              borderTopLeftRadius: 0,
              borderBottomLeftRadius: 0,
            },
          }}
        >
          <MenuItem value='minute'>Minute(s)</MenuItem>
          <MenuItem value='hour'>Hour(s)</MenuItem>
          <MenuItem value='day'>Day(s) (24h)</MenuItem>
          <MenuItem value='week'>Week(s)</MenuItem>
        </Select>
      </Grid>
    </Grid>
  )
}
