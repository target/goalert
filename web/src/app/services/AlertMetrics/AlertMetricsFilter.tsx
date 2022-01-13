import {
  Grid,
  Select,
  InputLabel,
  MenuItem,
  FormControl,
  SelectChangeEvent,
} from '@mui/material'
import { DateTime } from 'luxon'
import React from 'react'
import { useURLParam } from '../../actions/hooks'

interface AlertMetricsFilterProps {
  now: DateTime
}

export const MAX_DAY_COUNT = 28
export const DATE_FORMAT = 'y-MM-dd'

export default function AlertMetricsFilter({
  now,
}: AlertMetricsFilterProps): JSX.Element {
  const [since, setSince] = useURLParam<string>('since', '')

  const dateRangeValue = since
    ? Math.floor(
        now.diff(
          DateTime.fromFormat(since, DATE_FORMAT).minus({ day: 1 }),
          'weeks',
        ).weeks,
      )
    : MAX_DAY_COUNT / 7 // default

  const handleDateRangeChange = (e: SelectChangeEvent<number>): void => {
    const weeks = e.target.value as number
    setSince(
      now
        .minus({ weeks })
        .plus({ days: 1 })
        .startOf('day')
        .toFormat(DATE_FORMAT),
    )
  }

  return (
    <Grid container sx={{ marginLeft: '3rem' }}>
      <Grid item xs={5}>
        <FormControl sx={{ width: '100%' }}>
          <InputLabel id='demo-simple-select-helper-label'>
            Date Range
          </InputLabel>
          <Select
            fullWidth
            labelId='demo-simple-select-helper-label'
            id='demo-simple-select-helper'
            value={dateRangeValue}
            label='Date Range'
            name='date-range'
            onChange={handleDateRangeChange}
          >
            <MenuItem value={1}>Past week</MenuItem>
            <MenuItem value={2}>Past 2 weeks</MenuItem>
            <MenuItem value={3}>Past 3 weeks</MenuItem>
            <MenuItem value={4}>Past 4 weeks</MenuItem>
          </Select>
        </FormControl>
      </Grid>
    </Grid>
  )
}
