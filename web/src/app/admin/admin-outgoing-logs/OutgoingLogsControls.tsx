import React, { useState } from 'react'
import { Grid, Typography, IconButton } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import RestartAltIcon from '@mui/icons-material/RestartAlt'
import { theme } from '../../mui'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { useResetURLParams, useURLParam } from '../../actions'
import Search from '../../util/Search'
import { LOAD_AMOUNT } from './OutgoingLogsList'

interface Props {
  totalCount: number
}

const useStyles = makeStyles<typeof theme>({
  filterContainer: {
    display: 'flex',
    flexDirection: 'row',
  },
  resetButton: {
    height: 'min-content',
    alignSelf: 'center',
  },
})

export default function OutgoingLogsControls(p: Props): JSX.Element {
  const classes = useStyles()

  const [start, setStart] = useURLParam<string>('start', '')
  const [end, setEnd] = useURLParam<string>('end', '')
  const [key, setKey] = useState(0)
  const resetDateRange = useResetURLParams('start', 'end')
  const [limit] = useURLParam<string>('limit', '1')
  const _limit = parseInt(limit, 10)

  const resetFilters = (): void => {
    resetDateRange()
    setKey(key + 1)
  }

  return (
    <Grid container spacing={2} key={key}>
      <Grid item direction='column'>
        <Grid item>
          <ISODateTimePicker
            placeholder='Start'
            name='startDate'
            value={start}
            onChange={(newVal) => setStart(newVal as string)}
            label='Created after'
            margin='dense'
            size='small'
            variant='filled'
          />
        </Grid>
        <Grid item>
          <ISODateTimePicker
            placeholder='End'
            name='endDate'
            value={end}
            label='Created before'
            onChange={(newVal) => setEnd(newVal as string)}
            margin='dense'
            size='small'
            variant='filled'
          />
        </Grid>
      </Grid>
      <Grid item xs={1} sx={{ display: 'flex' }}>
        <IconButton
          className={classes.resetButton}
          type='button'
          onClick={resetFilters}
        >
          <RestartAltIcon />
        </IconButton>
      </Grid>
      <Grid item sx={{ flex: 1 }} />
      <Grid
        container
        item
        direction='column'
        justifyContent='flex-end'
        alignItems='flex-end'
        sx={{ width: 'fit-content' }}
      >
        <Grid item>
          <Search />
        </Grid>
        <Grid item>
          <Typography color='textSecondary'>
            Showing {_limit * LOAD_AMOUNT} of {p.totalCount} results
          </Typography>
        </Grid>
      </Grid>
    </Grid>
  )
}
