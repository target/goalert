import React from 'react'
import { Grid, Typography, IconButton } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import RestartAltIcon from '@mui/icons-material/RestartAlt'
import { ISODateTimePicker } from '../../util/ISOPickers'
import Search from '../../util/Search'
import { MAX_QUERY_ITEMS_COUNT } from './AdminDebugMessagesLayout'
import { useURLParams } from '../../actions'

interface Props {
  resetCount: () => void
  displayedCount: number
  resultsCount: number
}

const useStyles = makeStyles((theme: Theme) => {
  return {
    filterContainer: {
      display: 'flex',
      flexDirection: 'row',
    },
    resetButton: {
      height: 'min-content',
      alignSelf: 'center',
    },
    card: {
      padding: theme.spacing(2),
    },
    cardHeader: {
      padding: 0,
      paddingBottom: '1em',
    },
  }
})

export default function DebugMessagesControls({
  resetCount,
  displayedCount,
  resultsCount,
}: Props): JSX.Element {
  const classes = useStyles()

  const [params, setParams] = useURLParams({
    search: '',
    start: '',
    end: '',
  })

  const totalResultsCount =
    resultsCount < MAX_QUERY_ITEMS_COUNT
      ? resultsCount
      : `${MAX_QUERY_ITEMS_COUNT}+`

  return (
    <Card className={classes.card}>
      <CardHeader
        title='Outgoing Message Logs'
        className={classes.cardHeader}
      />
      <Grid container spacing={2}>
        <Grid item>
          <ISODateTimePicker
            placeholder='Start'
            name='startDate'
            value={params.start}
            onChange={(newStart) => {
              setParams({ ...params, start: newStart as string })
              resetCount()
            }}
            label='Created after'
            margin='dense'
            size='small'
            variant='outlined'
          />
        </Grid>
        <Grid item>
          <ISODateTimePicker
            placeholder='End'
            name='endDate'
            value={params.end}
            label='Created before'
            onChange={(newEnd) => {
              setParams({ ...params, end: newEnd as string })
              resetCount()
            }}
            margin='dense'
            size='small'
            variant='outlined'
          />
        </Grid>
      </Grid>
      <Grid item xs={1} sx={{ display: 'flex' }}>
        <IconButton
          className={classes.resetButton}
          type='button'
          onClick={resetCount}
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
            {`Showing ${displayedCount} of ${totalResultsCount} results`}
          </Typography>
        </Grid>
      </Grid>
    </Card>
  )
}
