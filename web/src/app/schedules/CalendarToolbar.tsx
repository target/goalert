import React from 'react'
import {
  Button,
  ButtonGroup,
  Grid,
  makeStyles,
  Typography,
} from '@material-ui/core'
import { DateTime } from 'luxon'
import { useURLParam } from '../actions'
import { getEndOfWeek, getStartOfWeek } from '../util/luxon-helpers'

const useStyles = makeStyles((theme) => ({
  container: {
    paddingBottom: '1em',
  },
  labelGridItem: {
    alignItems: 'center',
    display: 'flex',
    justifyContent: 'center',
    order: 3,
    [theme.breakpoints.up('lg')]: {
      order: 2,
    },
  },
  primaryNavBtnGroup: {
    flex: 1,
    display: 'flex',
    justifyContent: 'flex-start',
    order: 1,
  },
  secondaryBtnGroup: {
    display: 'flex',
    justifyContent: 'flex-start',
    order: 2,
    [theme.breakpoints.up('lg')]: {
      justifyContent: 'flex-end',
      order: 3,
    },
  },
}))

type ViewType = 'month' | 'week'
interface CalendarToolbarProps {
  startAdornment?: React.ReactNode
  endAdornment?: React.ReactNode
}

function CalendarToolbar(props: CalendarToolbarProps): JSX.Element {
  const classes = useStyles()
  const [weekly, setWeekly] = useURLParam<boolean>('weekly', false)
  const [start, setStart] = useURLParam(
    'start',
    weekly
      ? getStartOfWeek().toUTC().toISO()
      : DateTime.local().startOf('month').toUTC().toISO(),
  )

  const getLabel = (): string => {
    if (weekly) {
      const begin = getStartOfWeek(DateTime.fromISO(start))
        .toLocal()
        .toFormat('LLLL d')
      const end = getEndOfWeek(DateTime.fromISO(start))
        .toLocal()
        .plus({ days: 6 })
        .toFormat('LLLL d')
      return `${begin} — ${end}`
    }

    return DateTime.fromISO(start).toLocal().toFormat('LLLL yyyy')
  }

  /*
   * Resets the start date to the beginning of the month
   * when switching views.
   *
   * e.g. Monthly: February -> Weekly: Start at the week
   * of February 1st
   *
   * e.g. Weekly: February 17-23 -> Monthly: Start at the
   * beginning of February
   *
   * If viewing the current month however, show the current
   * week.
   */
  const onView = (nextView: ViewType): void => {
    const prevStartMonth = DateTime.fromISO(start).toLocal().month
    const currMonth = DateTime.local().month

    // if viewing the current month, show the current week
    if (nextView === 'week' && prevStartMonth === currMonth) {
      setWeekly(true)
      setStart(getStartOfWeek().toUTC().toISO())

      // if not on the current month, show the first week of the month
    } else if (nextView === 'week' && prevStartMonth !== currMonth) {
      setWeekly(true)
      setStart(
        DateTime.fromISO(start).toLocal().startOf('month').toUTC().toISO(),
      )

      // go from week to monthly view
      // e.g. if navigating to an overlap of two months such as
      // Jan 27 - Feb 2, show the latter month (February)
    } else {
      setWeekly(false)

      setStart(
        getEndOfWeek(DateTime.fromISO(start))
          .toLocal()
          .startOf('month')
          .toUTC()
          .toISO(),
      )
    }
  }

  const onNavigate = (next: DateTime): void => {
    if (weekly) {
      setStart(getStartOfWeek(next).toUTC().toISO())
    } else {
      setStart(next.toLocal().startOf('month').toUTC().toISO())
    }
  }

  const handleTodayClick = (): void => {
    onNavigate(DateTime.local())
  }

  const handleNextClick = (): void => {
    const timeUnit = weekly ? { weeks: 1 } : { months: 1 }
    const next = DateTime.fromISO(start).plus(timeUnit)
    onNavigate(next)
  }

  const handleBackClick = (): void => {
    const timeUnit = weekly ? { weeks: 1 } : { months: 1 }
    const next = DateTime.fromISO(start).minus(timeUnit)
    onNavigate(next)
  }

  return (
    <Grid container spacing={2} className={classes.container}>
      <Grid item xs={12} lg={4} className={classes.primaryNavBtnGroup}>
        {props.startAdornment}
        <ButtonGroup color='primary' aria-label='Calendar Navigation'>
          <Button data-cy='show-today' onClick={handleTodayClick}>
            Today
          </Button>
          <Button data-cy='back' onClick={handleBackClick}>
            Back
          </Button>
          <Button data-cy='next' onClick={handleNextClick}>
            Next
          </Button>
        </ButtonGroup>
      </Grid>

      <Grid item xs={12} lg={4} className={classes.labelGridItem}>
        <Typography component='p' data-cy='calendar-header' variant='subtitle1'>
          {getLabel()}
        </Typography>
      </Grid>

      <Grid item xs={12} lg={4} className={classes.secondaryBtnGroup}>
        <ButtonGroup
          color='primary'
          aria-label='Toggle between Monthly and Weekly views'
        >
          <Button
            data-cy='show-month'
            disabled={!weekly}
            onClick={() => onView('month')}
          >
            Month
          </Button>
          <Button
            data-cy='show-week'
            disabled={weekly}
            onClick={() => onView('week')}
          >
            Week
          </Button>
        </ButtonGroup>
        {props.endAdornment}
      </Grid>
    </Grid>
  )
}

export default CalendarToolbar
