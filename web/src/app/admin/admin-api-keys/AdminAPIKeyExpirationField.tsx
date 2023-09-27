import React, { useState, useEffect } from 'react'
import { FormContainer } from '../../forms'
import InputLabel from '@mui/material/InputLabel'
import MenuItem from '@mui/material/MenuItem'
import FormControl from '@mui/material/FormControl'
import Select, { SelectChangeEvent } from '@mui/material/Select'
import Grid from '@mui/material/Grid'
import { Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { ISODatePicker } from '../../util/ISOPickers'
import { DateTime } from 'luxon'

const useStyles = makeStyles(() => ({
  expiresCon: {
    'padding-top': '15px',
  },
}))

interface FieldProps {
  setValue: (val: string) => void
  value: string
  create: boolean
}

export default function AdminAPIKeyExpirationField(
  props: FieldProps,
): JSX.Element {
  const classes = useStyles()
  const { value, setValue } = props
  const [dateVal, setDateVal] = useState<string>(value)
  const [options, setOptions] = useState('7')
  const [showPicker, setShowPicker] = useState(false)
  // const today = DateTime.local({ zone: 'local' }).toFormat('ZZZZ')
  const handleChange = (event: SelectChangeEvent): void => {
    const val = event.target.value as string
    setOptions(val)
    setShowPicker(val.toString() === '0')

    if (val !== '0') {
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

  const handleDatePickerChange = (val: string): void => {
    // eslint-disable-next-line prettier/prettier
    if(val != null) {       
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
              disabled={!props.create}
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
              <ISODatePicker
                value={new Date(dateVal).toISOString()}
                onChange={handleDatePickerChange}
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
              The token will expires on {dateVal}
            </Typography>
          </Grid>
        </Grid>
      </FormControl>
    </FormContainer>
  )
}
