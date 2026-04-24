import React, { useEffect, useState } from 'react'
import InputLabel from '@mui/material/InputLabel'
import MenuItem from '@mui/material/MenuItem'
import Select from '@mui/material/Select'
import Grid from '@mui/material/Grid'
import { Typography } from '@mui/material'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { DateTime } from 'luxon'
import { Time } from '../../util/Time'
import { selectedDaysUntilTimestamp } from './util'

// props object for this component
interface FieldProps {
  onChange: (val: string) => void
  value: string
  disabled: boolean
  label: string
}

const presets = [7, 15, 30, 60, 90]

export default function AdminAPIKeyExpirationField(
  props: FieldProps,
): React.JSX.Element {
  const [selected, setSelected] = useState<number>(
    selectedDaysUntilTimestamp(props.value, presets),
  )

  useEffect(() => {
    if (!selected) return // if custom is selected, do nothing

    // otherwise keep the selected preset in sync with expiration time
    setSelected(selectedDaysUntilTimestamp(props.value, presets))
  }, [props.value])

  if (props.disabled) {
    return (
      <ISODateTimePicker
        disabled
        value={props.value}
        label={props.label}
        onChange={props.onChange}
      />
    )
  }

  return (
    <Grid container spacing={2}>
      <Grid justifyContent='flex-start' width='25%'>
        <InputLabel id='expires-at-select-label'>{props.label}</InputLabel>
        <Select
          label={props.label}
          fullWidth
          onChange={(e) => {
            const value = +e.target.value
            setSelected(value)

            if (value === 0) return
            props.onChange(DateTime.utc().plus({ days: value }).toISO())
          }}
          value={selected}
        >
          {presets.map((days) => (
            <MenuItem key={days} value={days}>
              {days} days
            </MenuItem>
          ))}
          <MenuItem value={0}>Custom...</MenuItem>
        </Select>
      </Grid>

      {selected ? ( // if a preset is selected, show the expiration time
        <Grid justifyContent='flex-start'>
          <Typography
            gutterBottom
            variant='subtitle2'
            component='div'
            sx={{ pt: '15px' }}
          >
            The token will expire <Time time={props.value} />
          </Typography>
        </Grid>
      ) : (
        // if custom is selected, show date picker
        <Grid justifyContent='flex-end' width='75%'>
          <ISODateTimePicker
            value={props.value}
            onChange={props.onChange}
            fullWidth
          />
        </Grid>
      )}
    </Grid>
  )
}
