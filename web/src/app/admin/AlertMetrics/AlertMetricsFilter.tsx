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
import { ServiceSelect } from '../../selection/ServiceSelect'

interface AlertMetricsFilterProps {
  now: DateTime
}

export const MAX_WEEKS_COUNT = 4

export default function AlertMetricsFilter(
  props: AlertMetricsFilterProps,
): JSX.Element {
  const [services, setServices] = useURLParam<string[]>('services', [])
  const [since, setSince] = useURLParam<string>('since', '')

  const dateRangeValue = since
    ? Math.floor(-DateTime.fromISO(since).diff(props.now, 'weeks').weeks)
    : MAX_WEEKS_COUNT // default

  const handleDateRangeChange = (e: SelectChangeEvent<number>): void => {
    const weeks = e?.target?.value as number
    setSince(props.now.minus({ weeks }).startOf('day').toISO())
  }

  return (
    <Grid container justifyContent='space-around'>
      <Grid item xs={5}>
        <ServiceSelect
          onChange={(v) => setServices(v)}
          multiple
          value={services}
          label='Filter by Service'
        />
      </Grid>
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
