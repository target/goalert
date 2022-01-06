import React, { useState } from 'react'
import { Grid, Typography, IconButton } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import RestartAltIcon from '@mui/icons-material/RestartAlt'
import { theme } from '../../mui'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { useResetURLParams, useURLParam } from '../../actions'
import Search from '../../util/Search'
import { MAX_QUERY_ITEMS_COUNT } from './AdminDebugMessagesLayout'
interface Props {
  showingLimit: number
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

export default function DebugMessagesControls(p: Props): JSX.Element {
  const classes = useStyles()

  const [start, setStart] = useURLParam<string>('start', '')
  const [end, setEnd] = useURLParam<string>('end', '')
  const [key, setKey] = useState(0)
  const resetDateRange = useResetURLParams('start', 'end')

  const resetFilters = (): void => {
    resetDateRange()
    // The ISODateTimePicker doesn't update to changes in it's `value` prop. It only uses it's internal state.
    // This key is a hotfix to set the ISODateTimePicker's value by just completely re-rendering it.
    setKey(key + 1)
  }

  const totalFetchedResultsCount =
    p.totalCount < MAX_QUERY_ITEMS_COUNT
      ? p.totalCount
      : `${MAX_QUERY_ITEMS_COUNT}+`

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
            {`Fetched ${Math.min(p.showingLimit, p.totalCount)} of
            ${totalFetchedResultsCount} results`}
          </Typography>
        </Grid>
      </Grid>
    </Grid>
  )
}
