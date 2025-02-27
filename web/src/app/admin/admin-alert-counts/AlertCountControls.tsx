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

export default function AlertCountControls(): React.JSX.Element {
  const now = useMemo(() => DateTime.now(), [])
  const [params, setParams] = useURLParams({
    createdAfter: now.minus({ days: 1 }).toISO(),
    createdBefore: '',
    interval: 'PT1H',
  })

  const handleFilterReset = useResetURLParams(
    'createdAfter',
    'createdBefore',
    'interval',
  )

  return (
    <Card>
      <Grid container spacing={1} sx={{ padding: 2 }}>
        <Grid item>
          <ISODateTimePicker
            sx={{ minWidth: '325px' }}
            placeholder='Start'
            name='startDate'
            value={params.createdAfter}
            size='small'
            onChange={(newStart) => {
              setParams({ ...params, createdAfter: newStart as string })
            }}
            label='Created After'
            variant='outlined'
          />
        </Grid>
        <Grid item>
          <ISODateTimePicker
            sx={{ minWidth: '325px' }}
            placeholder='End'
            name='endDate'
            value={params.createdBefore}
            label='Created Before'
            size='small'
            onChange={(newEnd) => {
              setParams({ ...params, createdBefore: newEnd as string })
            }}
            variant='outlined'
          />
        </Grid>
        <Grid item sx={{ flex: 1 }}>
          <FormControl sx={{ width: '100%', minWidth: '150px' }}>
            <InputLabel>Interval</InputLabel>
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
              <MenuItem value='PT1M'>Minute</MenuItem>
              <MenuItem value='PT1H'>Hour</MenuItem>
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
