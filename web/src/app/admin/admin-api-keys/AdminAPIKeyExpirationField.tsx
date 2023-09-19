import React, { useState, useEffect } from 'react'
import { FormContainer } from '../../forms'
import InputLabel from '@mui/material/InputLabel'
import MenuItem from '@mui/material/MenuItem'
import FormControl from '@mui/material/FormControl'
import Select, { SelectChangeEvent } from '@mui/material/Select'
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs'
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider'
import { DatePicker } from '@mui/x-date-pickers/DatePicker'
import dayjs, { Dayjs } from 'dayjs'
import Grid from '@mui/material/Grid'
import { Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'

const useStyles = makeStyles(() => ({
  expiresCon: {
    'padding-top': '15px',
  },
}))

interface FieldProps {
  setValue: (val: string) => void
  value: string
}

export default function AdminAPIKeyExpirationField(
  props: FieldProps,
): JSX.Element {
  const classes = useStyles()
  const { value, setValue } = props
  const [dateVal, setDateVal] = useState<string>(value)
  const [options, setOptions] = useState('7')
  const [showPicker, setShowPicker] = useState(false)
  const today = dayjs()
  const handleChange = (event: SelectChangeEvent): void => {
    const val = event.target.value as string
    setOptions(val)
    setShowPicker(val.toString() === '0')

    if (val !== '0') {
      setDateVal(today.add(parseInt(val), 'day').toString())
    }
  }

  const handleDatePickerChange = (val: Dayjs | null): void => {
    // eslint-disable-next-line prettier/prettier
    if(val != null) {
      setDateVal(val.toString())
    }
  }

  useEffect(() => {
    setValue(dateVal)
  })

  return (
    <FormContainer>
      <FormControl fullWidth>
        <Grid container spacing={2}>
          <Grid item justifyContent='flex-start'>
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
          {showPicker ? (
            <Grid item justifyContent='flex-end'>
              <LocalizationProvider dateAdapter={AdapterDayjs}>
                <DatePicker
                  value={dayjs(dateVal)}
                  minDate={today}
                  onChange={handleDatePickerChange}
                />
              </LocalizationProvider>
            </Grid>
          ) : null}
          <Grid item justifyContent='flex-start'>
            <Typography
              gutterBottom
              variant='subtitle2'
              component='div'
              className={classes.expiresCon}
            >
              The token will expires on {dateVal}
            </Typography>
          </Grid>
        </Grid>
      </FormControl>
    </FormContainer>
  )
}
