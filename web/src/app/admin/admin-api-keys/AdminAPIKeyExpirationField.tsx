import React, { useState, useEffect } from 'react'
import { FormContainer, FormField } from '../../forms'
import InputLabel from '@mui/material/InputLabel'
import MenuItem from '@mui/material/MenuItem'
import FormControl from '@mui/material/FormControl'
import Select, { SelectChangeEvent } from '@mui/material/Select'
import Grid from '@mui/material/Grid'
import { Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { ISODatePicker, ISODateTimePicker } from '../../util/ISOPickers'
import { DateTime } from 'luxon'
import { Time } from '../../util/Time'

const useStyles = makeStyles(() => ({
  expiresCon: {
    'padding-top': '15px',
  },
}))
// props object for this compoenent
interface FieldProps {
  setValue: (val: string) => void
  value: string
  disabled: boolean
}

export default function AdminAPIKeyExpirationField(
  props: FieldProps,
): JSX.Element {
  const classes = useStyles()
  const { value, setValue } = props
  const [dateVal, setDateVal] = useState<string>(value)
  const [options, setOptions] = useState('7')
  const [showPicker, setShowPicker] = useState(false)
  // handles expiration date field options changes: sets and computes expirates at value based on the selected additional days value or sets value to +7 days today if custom option is selected
  const handleChange = (event: SelectChangeEvent): void => {
    const val = event.target.value as string
    setOptions(val)
    setShowPicker(val.toString() === '0')

    if (val !== '0') {
      console.log(val)
      setDateVal(
        DateTime.now()
          .plus({ days: parseInt(val) })
          .toLocaleString({
            weekday: 'short',
            month: 'short',
            day: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
          }),
      )
    }
  }
  // handles custon expiration date field option changes: sets and computes expires at value based on the selected date
  const handleDatePickerChange = (val: string): void => {
    if (val === null) return

    setDateVal(
      new Date(val).toLocaleString([], {
        weekday: 'short',
        month: 'short',
        day: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      }),
    )
  }

  return (
    <FormContainer>
      <FormControl fullWidth>
        <Grid container spacing={2}>
          {props.disabled ? (
            <Grid item justifyContent='flex-start' width='100%'>
              <FormField
                fullWidth
                component={ISODateTimePicker}
                name='expiresAt'
                label='Expires At'
                value={new Date(dateVal).toLocaleString([], {
                  month: '2-digit',
                  day: '2-digit',
                  year: 'numeric',
                  hour: '2-digit',
                  minute: '2-digit',
                  second: '2-digit',
                })}
                required
                disabled
              />
            </Grid>
          ) : (
            <Grid item justifyContent='flex-start' width='25%'>
              <InputLabel id='expires-at-select-label'>Expires At*</InputLabel>
              <Select
                labelId='expires-at-select-label'
                id='expires-at-select'
                value={options}
                label='Expires At'
                onChange={handleChange}
                required
              >
                <MenuItem value={7}>7 days</MenuItem>
                <MenuItem value={15}>15 days</MenuItem>
                <MenuItem value={30}>30 days</MenuItem>
                <MenuItem value={60}>60 days</MenuItem>
                <MenuItem value={90}>90 days</MenuItem>
                <MenuItem value={0}>Custom...</MenuItem>
              </Select>
            </Grid>
          )}
          {showPicker ? (
            <Grid item justifyContent='flex-end' width='75%'>
              <ISODatePicker
                value={new Date(dateVal).toISOString()}
                onChange={handleDatePickerChange}
                fullWidth
              />
            </Grid>
          ) : null}
          <Grid item justifyContent='flex-start'>
            <Typography
              gutterBottom
              variant='subtitle2'
              component='div'
              className={classes.expiresCon}
            >
              The token will expires <Time time={value} />
            </Typography>
          </Grid>
        </Grid>
      </FormControl>
    </FormContainer>
  )
}
