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

export const MAX_WEEKS_COUNT = 4
export const DATE_FORMAT = 'y-MM-dd'

export default function AlertMetricsFilter(
  props: AlertMetricsFilterProps,
): JSX.Element {
  const [since, setSince] = useURLParam<string>('since', '')

  const dateRangeValue = since
    ? Math.floor(
        -DateTime.fromFormat(since, DATE_FORMAT).diff(props.now, 'weeks').weeks,
      )
    : MAX_WEEKS_COUNT // default

  const handleDateRangeChange = (e: SelectChangeEvent<number>): void => {
    const weeks = e.target.value as number
    setSince(props.now.minus({ weeks }).startOf('day').toFormat(DATE_FORMAT))
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
