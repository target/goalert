import {
  Grid,
  Select,
  InputLabel,
  MenuItem,
  FormControl,
  SelectChangeEvent,
} from '@mui/material'
import React from 'react'
import { useURLParam } from '../../actions/hooks'

export default function AlertMetricsFilter(): React.JSX.Element {
  const [range, setRange] = useURLParam<string>('range', 'P1M')
  const [ivl, setIvl] = useURLParam<string>('interval', 'P1D')

  return (
    <Grid container sx={{ marginLeft: '3rem', pt: 3 }}>
      <Grid item xs={5}>
        <FormControl sx={{ width: '100%' }}>
          <InputLabel id='demo-simple-select-helper-label'>
            Date Range
          </InputLabel>
          <Select
            fullWidth
            value={range}
            label='Date Range'
            name='date-range'
            onChange={(e: SelectChangeEvent<string>) =>
              setRange(e.target.value)
            }
          >
            <MenuItem value='P1W'>Past week</MenuItem>
            <MenuItem value='P2W'>Past 2 weeks</MenuItem>
            <MenuItem value='P1M'>Past Month</MenuItem>
            <MenuItem value='P3M'>Past 3 Months</MenuItem>
            <MenuItem value='P6M'>Past 6 Months</MenuItem>
            <MenuItem value='P1Y'>Past Year</MenuItem>
          </Select>
        </FormControl>
      </Grid>
      <Grid item xs={5} paddingLeft={1}>
        <FormControl sx={{ width: '100%' }}>
          <InputLabel id='demo-simple-select-helper-label'>Interval</InputLabel>
          <Select
            fullWidth
            value={ivl}
            label='Interval'
            name='interval'
            onChange={(e: SelectChangeEvent<string>) => setIvl(e.target.value)}
          >
            <MenuItem value='P1D'>Day</MenuItem>
            <MenuItem value='P1W'>Week</MenuItem>
            <MenuItem value='P1M'>Month</MenuItem>
          </Select>
        </FormControl>
      </Grid>
    </Grid>
  )
}
