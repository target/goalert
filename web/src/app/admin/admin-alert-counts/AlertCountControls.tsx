import React, { useMemo } from 'react'
import {
  Button,
  Grid,
  Select,
  InputLabel,
  MenuItem,
  FormControl,
  SelectChangeEvent,
  Card,
} from '@mui/material'
import ResetIcon from '@mui/icons-material/Replay'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { useURLParams, useResetURLParams } from '../../actions'
import { DateTime } from 'luxon'

export default function AlertCountControls(): JSX.Element {
  const now = useMemo(() => DateTime.now(), [])
  const [params, setParams] = useURLParams({
    since: now.minus({ months: 1 }).toISO(),
    until: now.toISO(),
    interval: 'P1D',
  })

  const handleFilterReset = useResetURLParams('since', 'until', 'interval')

  return (
    <Card>
      <Grid container spacing={1} sx={{ padding: 2 }}>
        <Grid item>
          <ISODateTimePicker
            placeholder='Start'
            name='startDate'
            value={params.since}
            size='small'
            onChange={(newStart) => {
              setParams({ ...params, since: newStart as string })
            }}
            label='Created After'
            variant='outlined'
          />
        </Grid>
        <Grid item>
          <ISODateTimePicker
            placeholder='End'
            name='endDate'
            value={params.until}
            label='Created Before'
            size='small'
            onChange={(newEnd) => {
              setParams({ ...params, until: newEnd as string })
            }}
            variant='outlined'
          />
        </Grid>
        <Grid item sx={{ flex: 1 }}>
          <FormControl sx={{ width: '100%' }}>
            <InputLabel id='demo-simple-select-helper-label'>
              Interval
            </InputLabel>
            <Select
              fullWidth
              value={params.interval}
              label='Interval'
              name='interval'
              size='small'
              onChange={(newInterval: SelectChangeEvent<string>) =>
                setParams({ ...params, interval: newInterval.target.value })
              }
            >
              <MenuItem value='P1H'>Hour</MenuItem>
              <MenuItem value='P1D'>Day</MenuItem>
              <MenuItem value='P1W'>Week</MenuItem>
              <MenuItem value='P1M'>Month</MenuItem>
            </Select>
          </FormControl>
        </Grid>
        <Grid item>
          <Button
            aria-label='Reset Filters'
            variant='outlined'
            onClick={() => handleFilterReset()}
            endIcon={<ResetIcon />}
            sx={{ height: '100%' }}
          >
            Reset
          </Button>
        </Grid>
      </Grid>
    </Card>
  )
}
