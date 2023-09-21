import React from 'react'
import { Button, Card, Grid } from '@mui/material'
import ResetIcon from '@mui/icons-material/Replay'
import { ISODateTimePicker } from '../../util/ISOPickers'
import Search from '../../util/Search'
import { useMessageLogsParams } from './util'

export default function AdminMessageLogsControls(): JSX.Element {
  const [params, setParams] = useMessageLogsParams()

  return (
    <Card>
      <Grid container spacing={1} sx={{ padding: 2 }}>
        <Grid item sx={{ flex: 1 }}>
          <Search transition={false} fullWidth />
        </Grid>
        <Grid item>
          <ISODateTimePicker
            placeholder='Start'
            name='startDate'
            value={params.start}
            onChange={(newStart) => {
              setParams({ ...params, start: newStart as string })
            }}
            label='Created After'
            size='small'
            variant='outlined'
          />
        </Grid>
        <Grid item>
          {/* TODO: allow temporary invalid state if typing.
        
          e.g. start: 9/14. with an end set at 9/30, the user wants to
          set end to 9/15 by typing "1" then "5" in the day field.
          This currently results in a no-op as "9/1" is less than "9/14"
         */}
          <ISODateTimePicker
            placeholder='End'
            name='endDate'
            value={params.end}
            label='Created Before'
            onChange={(newEnd) => {
              setParams({ ...params, end: newEnd as string })
            }}
            size='small'
            variant='outlined'
          />
        </Grid>
        <Grid item>
          <Button
            aria-label='Reset Filters'
            variant='outlined'
            onClick={() => {
              setParams({
                search: '',
                start: '',
                end: '',
              })
            }}
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
