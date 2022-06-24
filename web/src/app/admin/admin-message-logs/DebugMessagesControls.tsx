import React from 'react'
import { Button, Card, CardActions, CardContent, Grid } from '@mui/material'
import ResetIcon from '@mui/icons-material/Replay'
import { ISODateTimePicker } from '../../util/ISOPickers'
import Search from '../../util/Search'
import { useURLParams } from '../../actions'

interface Props {
  resetCount: () => void
}

export default function DebugMessagesControls({
  resetCount,
}: Props): JSX.Element {
  const [params, setParams] = useURLParams({
    search: '',
    start: '',
    end: '',
  })

  // const totalResultsCount =
  //   resultsCount < MAX_QUERY_ITEMS_COUNT
  //     ? resultsCount
  //     : `${MAX_QUERY_ITEMS_COUNT}+`

  return (
    <Card>
      <CardContent>
        <Grid container direction='column' spacing={1}>
          <Grid item>
            <ISODateTimePicker
              placeholder='Start'
              name='startDate'
              value={params.start}
              onChange={(newStart) => {
                setParams({ ...params, start: newStart as string })
                resetCount()
              }}
              label='Created After'
              margin='dense'
              size='small'
              variant='outlined'
              fullWidth
            />
          </Grid>
          <Grid item>
            <ISODateTimePicker
              placeholder='End'
              name='endDate'
              value={params.end}
              label='Created Before'
              onChange={(newEnd) => {
                setParams({ ...params, end: newEnd as string })
                resetCount()
              }}
              margin='dense'
              size='small'
              variant='outlined'
              fullWidth
            />
          </Grid>
          <Grid item sx={{ flex: 1 }} />
          <Grid item>
            <Search transition={false} fullWidth />
          </Grid>
        </Grid>
      </CardContent>
      <CardActions sx={{ p: 2 }}>
        <Button
          aria-label='Reset Filters'
          variant='outlined'
          onClick={resetCount}
          startIcon={<ResetIcon />}
        >
          Reset
        </Button>
      </CardActions>
    </Card>
  )
}
